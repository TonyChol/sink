package fs

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
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

// TraverseDir : A wrapper function that returns
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

// UpdateFileDB : A filepath.WalkFunc function that updates the db
// whenever it meets a file in the synching directory
func updateFileDB(fileDB *FileDB) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// 1. [x] get file type
		// 2. [x] get file mode
		// 3. [x] get checksum of the file if it is not directory
		// 4. [x] get the last modify
		var checkSum string
		if info.IsDir() {
			checkSum = ""
		} else {
			checkSum, err = getCheckSumOfFile(path)
			if err != nil {
				log.Fatal("Checksum error", err)
				return err
			}
		}

		fileDBEle := FileDBElement{}
		fileDBEle.FileType = getFileType(info)
		fileDBEle.Mode = getFileMode(info)
		fileDBEle.CheckSum = checkSum
		fileDBEle.LastModify = info.ModTime()
		(*fileDB)[path] = fileDBEle

		return nil
	}
}

// ScanDir : Scan the whole directory to update the file database
// and store some information
func ScanDir(dirPath string) *FileDB {
	filedb := GetFileDBInstance()
	filepath.Walk(dirPath, updateFileDB(filedb))
	return filedb
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

// getFileType : Get type according to the file info
func getFileType(info os.FileInfo) string {
	if info.IsDir() {
		return "d"
	}

	return "f"
}

// getFileMode : Get the fileMode string of the file
// A FileMode represents a file's mode and permission bits.
// The bits have the same definition on all systems, so that
// information about files can be moved from one system
// to another portably. Not all bits apply to all systems.
func getFileMode(info os.FileInfo) os.FileMode {
	return info.Mode()
}

// getCheckSumOfFile : Get the checksum string of one file
// returns error if the filePath is not valid
func getCheckSumOfFile(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}
