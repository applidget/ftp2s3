package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	log.Info("Starting watcher ...")

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

	log.Infof("Watching directory %s", workingDir)
	if err := setupRecursiveWatch(workingDir, watcher); err != nil {
		log.Fatal(err)
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
						log.Error(err)
						return
					}
					log.Info(url)
					if err := notifyNewImage("lol", url); err != nil {
						log.Error(err)
					}
				}()

			case inotify.IN_CREATE | inotify.IN_ISDIR: //new directory created inside working dir, watch it too
				if err := watcher.AddWatch(ev.Name, inotify.IN_ALL_EVENTS); err != nil {
					log.Error(err)
				} else {
					log.Infof("now watching %s", ev.Name)
					//add other sub directories
					if err := setupRecursiveWatch(ev.Name, watcher); err != nil {
						log.Fatal(err)
					}
				}

			case inotify.IN_DELETE | inotify.IN_ISDIR:
				if err := watcher.RemoveWatch(ev.Name); err != nil {
					log.Warn(err)
				} else {
					log.Infof("now watching %s", ev.Name)
				}
			}

		case err := <-watcher.Error:
			log.Error(err)
		}
	}
}

// setupRecursiveWatch recursively watch folders from the given path (given path is supposed to be watched already)

func setupRecursiveWatch(basePath string, watcher *inotify.Watcher) error {
	walkFn := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		absPath, _ := filepath.Abs(p)
		log.Infof("- watching %s", absPath)
		return watcher.AddWatch(p, inotify.IN_ALL_EVENTS)
	}
	return filepath.Walk(basePath, walkFn)
}

// uploadImageToS3 check if the given file is an image and upload it to S3 and post to web_hook if . It return the
// image URL or an error
func uploadImageToS3(p string) (string, error) {
	log.Infof("handling %s", p)

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
		return "", fmt.Errorf("only file with extensions %v are uploaded", allowedExts)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	if err = bucket.Put(fileName, b, filepath.Join("image", ext), "public-read-write"); err != nil {
		return "", err
	}

	log.Info("file %s uploaded to S3\n", p)
	os.Remove(p)
	return bucket.URL(fileName), nil
}

// notifyNewImage send a POST request to the WEB_HOOK env var (if it exists)

func notifyNewImage(basePath, imageUrl string) error {
	hook := os.Getenv("WEB_HOOK")
	if hook == "" {
		return nil
	}

	type payload struct {
		Url      string `json:"remote_photo_url"`
		BasePath string `json:"base_path"`
	}

	body := &payload{Url: imageUrl, BasePath: basePath}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", hook, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("expecting status code in 200 .. 299 got %d", resp.StatusCode)
	}
	log.Infof("Notification sent to %s", hook)
	return nil
}
