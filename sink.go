package main

import (
	"errors"
	"io/ioutil"
	"log"

	"os"

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
	defer watcher.RemoveWatch(dir)

	if err != nil {
		log.Fatal(err)
		return
	}

	files, err := ioutil.ReadDir("./testDir")

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		log.Println(file.Name())
		if file.IsDir() == true {
			watchDir(file.Name())
		}
	}
	<-done
}

func getAbsolutePath() (string, error) {
	return os.Getwd()
}

func getDirPathFromAgrs() (string, error) {
	argsArr := os.Args
	if len(argsArr) < 2 {
		err := errors.New("You should attach a file directory")
		return "", err
	}
	return os.Args[1], nil
}

func main() {
	rootDir, err := getAbsolutePath()
	if err != nil {
		log.Fatal(err)
	}

	relativeDir, err := getDirPathFromAgrs()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Args[1]: ", relativeDir)

	log.Print("Root Dir: ", rootDir)

	watchDir("testDir")
}
