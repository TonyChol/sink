package main

import (
	"container/list"
	"errors"
	"log"
	"os"
	"path/filepath"

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

	// files, err := ioutil.ReadDir("./testDir")

	if err != nil {
		log.Fatal(err)
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

func traverseDir(fl *list.List) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if info.IsDir() {
			fl.PushBack(path)
		}

		return nil
	}
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

	targetDir := rootDir + "/" + relativeDir

	log.Println("target directory: ", targetDir)
	log.Print("Root Dir: ", rootDir)

	l := list.New()
	filepath.Walk(targetDir, traverseDir(l))

	log.Println("list's length = ", l.Len())

	for e := l.Front(); e != nil; e = e.Next() {
		folder := e.Value
		go watchDir(folder.(string))
	}

	log.Println("Start setting up file watcher for each directory in ", rootDir)

	done := make(chan bool)
	<-done
}
