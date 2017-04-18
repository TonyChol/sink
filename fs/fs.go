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

// Base64StrFromFile : Encode the file into the base64 string
func Base64StrFromFile(path string) (string, error) {
	buff, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buff), nil

}

// FileFromBase64Str : Decode the base64 string into a file
func FileFromBase64Str(code string, dest string) error {
	buff, err := base64.StdEncoding.DecodeString(code)

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dest, buff, 07400); err != nil {
		return err
	}

	return nil
}
