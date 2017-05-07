package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"

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
func watchDir(done chan bool, deviceID string, baseDir string, dirs ...string) {
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

					if fi.IsDir() == false {
						_, exist := (*fs.GetFileDBInstance())[eventFile]
						if exist == false {
							fs.GetFileDBInstance().AddFileDir(eventFile)
						}
					}
				}

				if ev.IsDelete() {
					watcher.RemoveWatch(eventFile)
					// TODO: let server delete the file
					// requestDel(eventFile)
				}

				if ev.IsModify() {
				}

				if ev.IsRename() {
				}

				if ev.IsAttrib() {
					continue
				}

				relativePath, err := filepath.Rel(baseDir, path.Dir(eventFile))
				if err != nil {
					log.Fatalln("can not get relative path of the file", eventFile)
				}

				fileDb := fs.GetFileDBInstance()
				_, exist := (*fileDb)[eventFile]

				if exist == false {
					log.Println("doesn't exist")
					err = sync.SendFile(eventFile, relativePath, deviceID)
					if err != nil {
						log.Printf("Can not send file %v: %v", eventFile, err)
					}
				} else {
					ele := (*fileDb)[eventFile]
					if ele.Incoming == false {
						log.Println("incoming is false", eventFile)
						err = sync.SendFile(eventFile, relativePath, deviceID)
						if err != nil {
							log.Printf("Can not send file %v: %v", eventFile, err)
						}
					} else {
						// set Incoming to false
						fs.GetFileDBInstance().UnsetIncoming(eventFile)
					}
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

func getDeviceIDFromArgs() (string, error) {
	argsArr := os.Args
	if len(argsArr) < 3 {
		err := errors.New("You should attach an unique device id")
		return "", err
	}
	return argsArr[2], nil
}

func main() {
	rootDir, err := fs.GetAbsolutePath()
	util.HandleErr(err)

	relativeDir, err := fs.GetDirPathFromAgrs()
	util.HandleErr(err)

	deviceID, err := getDeviceIDFromArgs()
	util.HardHandleErr(err)

	getFileDB()

	targetDir := rootDir + "/" + relativeDir

	// Do the scan for the first time
	filedb := scanner.ScanDir(targetDir, deviceID)
	filedb.SaveDBAsJSON()

	dirSlice := fs.AllRecursiveDirsIn(targetDir)

	// start firing the file watcher
	done := make(chan bool)
	go watchDir(done, deviceID, targetDir, dirSlice...)
	log.Println("Start setting up file watcher for each directory in ", targetDir)
	// launch socket connection
	go getFreePortAndConnect(targetDir, deviceID)

	// wait for exit signal
	// reference: http://stackoverflow.com/questions/8403862/do-actions-on-end-of-execution
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	done <- true
	log.Println("\nSink program got killed !")
	os.Exit(0)
}

func getFreePortAndConnect(targetDir string, deviceID string) {
	var res = &(networking.PortPayload{})
	conf := config.GetInstance()
	// remoteAddr := config.GetInstance().DevServer + fmt.Sprintf(":%d", config.GetInstance().DevPort) + "/socketPort"
	remoteAddr := "http://" + conf.DevServer + fmt.Sprintf(":%d", conf.DevPort) + conf.FreeSocketPattern
	// Get available port for socket connection from server
	networking.GetJSON(remoteAddr, res)

	// start accepting files
	socketAddr := config.GetInstance().DevServer + fmt.Sprintf(":%d", res.Data.Port)
	sync.ConnectSocket(socketAddr, targetDir, deviceID)
}
