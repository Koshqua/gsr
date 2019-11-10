package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

func main() {
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()
	filepath.Walk(".", addListener)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				fmt.Println("EVENT: ", event)
			case err := <-watcher.Errors:
				fmt.Print("ERROR: ", err)
			}
		}
	}()
	<-done
}

func addListener(path string, fi os.FileInfo, err error) error {
	//checking if it's a .go file
	fname := strings.Split(fi.Name(), ".")
	format := ""
	if len(fname) > 1 {
		format = fname[1]
	}
	//adding watcher to to .go files in every folder
	if format == "go" {
		watcher.Add(path)
	}
	return nil
}
