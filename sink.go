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

// watchDir : A goroutine to watch all the directories
func watchDir(done chan bool, dirs ...string) {
	watcher, err := fsnotify.NewWatcher()
	util.HandleErr(err)
	defer watcher.Close()

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				eventFile := ev.Name
				if ev.IsCreate() {
					tempFile, err := os.Open(eventFile)
					util.HandleErr(err)

					fi, err := tempFile.Stat()
					util.HandleErr(err)
					switch {
					case fi.IsDir():
						log.Println("File", eventFile, "is created! Start watching this new folder")
						log.Println("Will start sending new file to other endpoints")
						log.Println()
						err = watcher.Watch(eventFile)
						if err != nil {
							log.Fatal(err)
							return
						}
					}
				}

				if ev.IsDelete() {
					log.Println("File", eventFile, "is deleted! Stop watching this new folder")
					log.Println("Will start notifying this deleted directory to other endpoints")
					log.Println()
					watcher.RemoveWatch(eventFile)
				}

				if ev.IsModify() {
					log.Println("File", eventFile, "is modified!")
					log.Println("Will start notifying this modified directory to other endpoints")
					log.Println()
				}

				if ev.IsRename() {
					log.Println("File", eventFile, "is renamed!")
					log.Println("Will start notifying this renamed directory to other endpoints")
					log.Println()
				}

				if ev.IsAttrib() {
					log.Println("File", eventFile, "'s attributes are changed")
					log.Println("Will start notifying that the attributes of this directory is changed")
					log.Println()
				}
			case err := <-watcher.Error:
				log.Println("error", err)
				done <- true
			}
		}
	}()

	for _, dir := range dirs {
		err = watcher.Watch(dir)
		defer watcher.RemoveWatch(dir)
		if err != nil {
			log.Fatal(err)
			done <- true
		}
	}

	<-done
}

// allRecursiveDirsIn : Get all the directories string inside the dirPath
func allRecursiveDirsIn(dirPath string) []string {
	l := list.New()

	filepath.Walk(dirPath, fs.TraverseDir(l))

	var dirSlice = make([]string, l.Len())

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		dirSlice[i] = e.Value.(string)
		i++
	}

	return dirSlice
}

func main() {

	rootDir, err := fs.GetAbsolutePath()
	util.HandleErr(err)

	relativeDir, err := fs.GetDirPathFromAgrs()
	util.HandleErr(err)

	targetDir := rootDir + "/" + relativeDir

	log.Println("target directory: ", targetDir)
	log.Print("Root Dir: ", rootDir)

	dirSlice := allRecursiveDirsIn(targetDir)

	done := make(chan bool)
	go watchDir(done, dirSlice...)

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	exit := make(chan bool)
	<-exit
	done <- true
}
