package fs

import (
	"container/list"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// GetAbsolutePath : Get the absolute path
// of the location that starts the program
func GetAbsolutePath() (string, error) {
	return os.Getwd()
}

// GetDirPathFromAgrs : Get the input path
// from the command-line arguments
func GetDirPathFromAgrs() (string, error) {
	argsArr := os.Args
	if len(argsArr) < 2 {
		err := errors.New("You should attach a file directory")
		return "", err
	}
	return os.Args[1], nil
}

// TraverseDir : A wrapper function that return
// the filepath.WalkFunc function which would be used by filepath.Walk
func TraverseDir(fl *list.List) filepath.WalkFunc {
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

// AllRecursiveDirsIn : Get all the directories string inside the dirPath
func AllRecursiveDirsIn(dirPath string) []string {
	l := list.New()

	filepath.Walk(dirPath, TraverseDir(l))

	var dirSlice = make([]string, l.Len())

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		dirSlice[i] = e.Value.(string)
		i++
	}

	return dirSlice
}

	}

	return base64.StdEncoding.EncodeToString(buff), nil

}


	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dest, buff, 07400); err != nil {
		return err
	}

	return nil
}
