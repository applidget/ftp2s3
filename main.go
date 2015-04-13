package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"golang.org/x/exp/inotify"
)

var (
	bucket      *s3.Bucket
	allowedExts = []string{"png", "jpeg", "jpg"}
)

func main() {
	log.Println("Starting watcher ...")

	//instanciate S3 client
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}
	client := s3.New(auth, aws.EUWest)
	bucket = client.Bucket("applidget-ftp-photo-uploader")

	//listen fo file creation on /tmp
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	dirToWatch := "./"
	if len(os.Args) > 1 {
		dirToWatch = os.Args[1]
	}

	log.Printf("Watching directory %s\n", dirToWatch)
	err = watcher.Watch(dirToWatch)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask == inotify.IN_CLOSE_WRITE { //if it's a file we handle it
				go fileCreated(ev.Name)
			}
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}

func fileCreated(p string) {
	log.Printf("handling %s\n", p)

	fileName := path.Base(p)
	ext := strings.TrimPrefix(path.Ext(p), ".")

	allowed := false
	for _, e := range allowedExts {
		allowed = e == ext
		if allowed {
			break
		}
	}

	if !allowed {
		log.Printf("only file with extensions are %v uploaded\n", allowedExts)
		return
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		log.Printf("unable to read file %s: %v\n", p, err)
		return
	}

	if err = bucket.Put(path.Join("raw", fileName), b, path.Join("image", ext), "public-read-write"); err != nil {
		log.Printf("unable to upload file %s: %v\n", p, err)
		return
	}

	log.Printf("file %s uploaded to S3\n", p)
	os.Remove(p)
}
