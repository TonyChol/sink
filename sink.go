package main

import (
	"container/list"
	"log"
	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/util"
)

func watchDir(dirs ...string) {
	watcher, err := fsnotify.NewWatcher()
	util.HandleErr(err)
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				if ev.IsCreate() {
					log.Println("Dir", ev.Name, "is created! Start watch this new folder")
					newDir := ev.Name
					err = watcher.Watch(newDir)
					if err != nil {
						log.Fatal(err)
						return
					}
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

	log.Println("list's length = ", l.Len())

	var dirSlice = make([]string, l.Len())

	idx := 0
	for e := l.Front(); e != nil; e = e.Next() {
		folder := e.Value
		dirSlice[idx] = folder.(string)
		idx++
	}

	go watchDir(dirSlice...)

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	done := make(chan bool)
	<-done
}
