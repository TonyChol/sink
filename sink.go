package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/howeyc/fsnotify"
	"github.com/tonychol/sink/config"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/networking"
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
				err := sync.SendFile(eventFile)
				if err != nil {
					log.Printf("Can not send file %v: %v", eventFile, err)
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

	// Do the scan for the first time
	filedb := scanner.ScanDir(targetDir)
	filedb.SaveDBAsJSON()

	dirSlice := fs.AllRecursiveDirsIn(targetDir)

	// start firing the file watcher
	done := make(chan bool)
	go watchDir(done, dirSlice...)
	log.Println("Start setting up file watcher for each directory in ", targetDir)
	// launch socket connection
	go getFreePort()

	// wait for exit signal
	// reference: http://stackoverflow.com/questions/8403862/do-actions-on-end-of-execution
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	done <- true
	log.Println("\nSink program got killed !")
	os.Exit(0)
}

func getFreePort() {
	var res = &(networking.PortPayload{})
	// remoteAddr := config.GetInstance().DevServer + fmt.Sprintf(":%d", config.GetInstance().DevPort) + "/socketPort"
	remoteAddr := config.GetInstance().DevServer + fmt.Sprintf(":%d", 8181) + config.GetInstance().FreeSocketPattern
	networking.GetJSON(remoteAddr, res)

	// start accepting files
	socketAddr := fmt.Sprintf("localhost:%d", res.Data.Port)
	sync.ConnectSocket(socketAddr)
}
