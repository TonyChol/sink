package main

import (
	"container/list"
	"log"
	"os"
	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/util"
)

func watchDir(done chan bool, dirs ...string) {
	watcher, err := fsnotify.NewWatcher()
	util.HandleErr(err)
	defer watcher.Close()

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				eventDir := ev.Name
				if ev.IsCreate() {
					tempFile, err := os.Open(eventDir)
					util.HandleErr(err)

					fi, err := tempFile.Stat()
					util.HandleErr(err)
					switch {
					case fi.IsDir():
						log.Println("File", eventDir, "is created! Start watching this new folder")
						log.Println("Will start sending new file to other endpoints")
						log.Println()
						err = watcher.Watch(eventDir)
						if err != nil {
							log.Fatal(err)
							return
						}
					}
				}

				if ev.IsDelete() {
					log.Println("File", eventDir, "is deleted! Stop watching this new folder")
					log.Println("Will start notifying this deleted directory to other endpoints")
					log.Println()
					watcher.RemoveWatch(eventDir)
				}

				if ev.IsModify() {
					log.Println("File", eventDir, "is modified!")
					log.Println("Will start notifying this modified directory to other endpoints")
					log.Println()
				}

				if ev.IsRename() {
					log.Println("File", eventDir, "is renamed!")
					log.Println("Will start notifying this renamed directory to other endpoints")
					log.Println()
				}

				if ev.IsAttrib() {
					log.Println("File", eventDir, "'s attributes are changed")
					log.Println("Will start notifying this renamed directory to other endpoints")
					log.Println()
				}
			case err := <-watcher.Error:
				log.Println("error", err)
			}
		}
	}()

	for _, dir := range dirs {
		err = watcher.Watch(dir)
		defer watcher.RemoveWatch(dir)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	<-done
}

func main() {

	rootDir, err := fs.GetAbsolutePath()
	util.HandleErr(err)

	relativeDir, err := fs.GetDirPathFromAgrs()
	util.HandleErr(err)

	targetDir := rootDir + "/" + relativeDir

	log.Println("target directory: ", targetDir)
	log.Print("Root Dir: ", rootDir)

	l := list.New()

	filepath.Walk(targetDir, fs.TraverseDir(l))

	var dirSlice = make([]string, l.Len())

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		dirSlice[i] = e.Value.(string)
		i++
	}

	done := make(chan bool)

	go watchDir(done, dirSlice...)

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	exit := make(chan bool)
	<-exit
	done <- true
}
