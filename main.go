package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"golang.org/x/exp/inotify"
)

var (
	bucket      *s3.Bucket
	allowedExts = []string{"png", "jpeg", "jpg"}
	workingDir  = "./" //can be overrided by args[1]
)

func main() {
	log.Println("Starting watcher ...")

	//instanciate S3 client
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}
	client := s3.New(auth, aws.EUWest)
	bucket = client.Bucket(os.Getenv("AWS_BUCKET"))

	//listen fo file creation on /tmp
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		workingDir = os.Args[1]
	}

	log.Printf("Watching directory %s\n", workingDir)
	if err := setupRecursiveWatch(workingDir, watcher); err != nil {
		log.Fatal("[ERROR in setup recursive watch] %v\n", err)
	}
	err = watcher.Watch(workingDir)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case ev := <-watcher.Event:
			switch ev.Mask {

			case inotify.IN_CLOSE_WRITE: //if it's a file we handle it
				go func() {
					url, err := uploadImageToS3(ev.Name)
					if err != nil {
						log.Printf("[ERROR uploading to S3] %v\n", err)
					} else {
						log.Println(url)
					}
				}()

			case inotify.IN_CREATE | inotify.IN_ISDIR: //new directory created inside working dir, watch it too
				if err := watcher.AddWatch(ev.Name, inotify.IN_ALL_EVENTS); err != nil {
					log.Printf("[ERROR adding watch on new folder] %v\n", err)
				} else {
					log.Printf("now watching %s\n", ev.Name)
				}

			case inotify.IN_DELETE | inotify.IN_ISDIR:
				if err := watcher.RemoveWatch(ev.Name); err != nil {
					log.Printf("[ERROR removing watch] %v\n", err)
				} else {
					log.Printf("now watching %s\n", ev.Name)
				}
			}

		case err := <-watcher.Error:
			log.Printf("[ERROR] %v\n", err)
		}
	}
}

// setupRecursiveWatch recursively watch folders from the given path (given path is supposed to be watched already)

func setupRecursiveWatch(basePath string, watcher *inotify.Watcher) error {
	walkFn := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		return watcher.AddWatch(p, inotify.IN_ALL_EVENTS)
	}
	return filepath.Walk(basePath, walkFn)
}

// uploadImageToS3 check if the given file is an image and upload it to S3. It return the
// image URL or an error
func uploadImageToS3(p string) (string, error) {
	log.Printf("handling %s\n", p)

	fileName, _ := filepath.Rel(workingDir, p)
	ext := strings.TrimPrefix(filepath.Ext(p), ".")

	allowed := false
	for _, e := range allowedExts {
		allowed = e == ext
		if allowed {
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("only file with extensions are %v uploaded\n", allowedExts)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	if err = bucket.Put(fileName, b, filepath.Join("image", ext), "public-read-write"); err != nil {
		return "", err
	}

	log.Printf("file %s uploaded to S3\n", p)
	os.Remove(p)
	return bucket.URL(fileName), nil
}
