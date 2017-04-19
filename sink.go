package main

import (
	"log"
	"os"

	"github.com/howeyc/fsnotify"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/util"
)

// watchDir : A goroutine to watch all the directories
// and fires the specific file event
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

func main() {

	rootDir, err := fs.GetAbsolutePath()
	util.HandleErr(err)

	relativeDir, err := fs.GetDirPathFromAgrs()
	util.HandleErr(err)

	targetDir := rootDir + "/" + relativeDir

	log.Println("target directory: ", targetDir)
	log.Print("Root Dir: ", rootDir)

	// Do the scan for the first time
	filedb := fs.ScanDir(targetDir)
	log.Println(filedb.JsonStr())

	dirSlice := fs.AllRecursiveDirsIn(targetDir)

	done := make(chan bool)
	go watchDir(done, dirSlice...) // start firing the file watcher

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	exit := make(chan bool)
	<-exit
	done <- true
}
