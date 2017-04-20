package main

import (
	"log"
	"os"

	"github.com/howeyc/fsnotify"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/scanner"
	"github.com/tonychol/sink/sync"
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
				eventFile := ev.Name
				log.Println("event:", ev)

				if ev.IsCreate() {
					tempFile, err := os.Open(eventFile)
					util.HandleErr(err)

					fi, err := tempFile.Stat()
					util.HandleErr(err)
					switch {
					case fi.IsDir():
						err = watcher.Watch(eventFile)
						if err != nil {
							log.Fatal(err)
							return
						}
					}
				}

				if ev.IsDelete() {
					watcher.RemoveWatch(eventFile)
				}

				if ev.IsModify() {
				}

				if ev.IsRename() {
				}

				if ev.IsAttrib() {
				}
				sync.SendFile(eventFile)
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

// getFileDB : At the beginning of the program, the db file
// that describes the synching directory is restored
func getFileDB() {
	log.Println("Restoring db instance from json file")
	_ = fs.GetFileDBInstance()
}

func main() {

	rootDir, err := fs.GetAbsolutePath()
	util.HandleErr(err)

	relativeDir, err := fs.GetDirPathFromAgrs()
	util.HandleErr(err)

	getFileDB()

	targetDir := rootDir + "/" + relativeDir
	log.Println("target directory: ", targetDir)
	log.Print("Root Dir: ", rootDir)

	// Do the scan for the first time
	filedb := scanner.ScanDir(targetDir)
	filedb.SaveDBAsJSON()

	dirSlice := fs.AllRecursiveDirsIn(targetDir)

	done := make(chan bool)
	go watchDir(done, dirSlice...) // start firing the file watcher

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	exit := make(chan bool)
	<-exit
	done <- true
}
