package main

import (
	"log"

	"github.com/howeyc/fsnotify"
)

func watchDir(dir string) {
	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()

	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event: ", ev)
			case err := <-watcher.Error:
				log.Println("error", err)
			}
		}
	}()

	err = watcher.Watch(dir)
	if err != nil {
		log.Fatal(err)
		return
	}

	<-done
}

func main() {
	watchDir("testDir")
}
