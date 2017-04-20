package scanner

import (
	"log"
	"os"
	"path/filepath"

	"github.com/tonychol/sink/fs"
	syncing "github.com/tonychol/sink/sync"
)

// UpdateFileDB : A filepath.WalkFunc function that updates the db
// whenever it meets a file in the synching directory
func updateFileDB(fileDB *fs.FileDB) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// 1. [x] get file type
		// 2. [x] get file mode
		// 3. [x] get checksum of the file if it is not directory
		// 4. [x] get the last modify
		var checkSum string
		if info.IsDir() {
			checkSum = ""
		} else {
			checkSum, err = fs.GetCheckSumOfFile(path)
			if err != nil {
				log.Fatal("Checksum error", err)
				return err
			}
		}

		fileDBEle := fs.FileDBElement{}
		fileDBEle.FileType = fs.GetFileType(info)
		fileDBEle.Mode = fs.GetFileMode(info)
		fileDBEle.CheckSum = checkSum
		fileDBEle.LastModify = info.ModTime()

		_, exist := (*fileDB)[path]
		// if is the file and the path does not exist in db
		if info.IsDir() == false && exist == false {
			log.Println("should send the file", path, "to server!")
			syncing.SendFile(path)
		}

		(*fileDB)[path] = fileDBEle

		return nil
	}
}

// ScanDir : Scan the whole directory to update the file database
// and store some information
func ScanDir(dirPath string) *fs.FileDB {
	filedb := fs.GetFileDBInstance()
	filepath.Walk(dirPath, updateFileDB(filedb))
	return filedb
}
