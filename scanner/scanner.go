package scanner

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/tonychol/sink/fs"
	syncing "github.com/tonychol/sink/sync"
)

// UpdateFileDB : A filepath.WalkFunc function that updates the db
// whenever it meets a file in the synching directory
func updateFileDB(fileDB *fs.FileDB, deviceID string, baseDir string) filepath.WalkFunc {
	return func(fpath string, info os.FileInfo, err error) error {
		var checkSum string
		if info.IsDir() {
			checkSum = ""
		} else {
			checkSum, err = fs.GetCheckSumOfFile(fpath)
			if err != nil {
				log.Fatal("Checksum error", err)
				return err
			}
		}

		fileDBEle := fs.NewFileDBEle()
		fileDBEle.FileType = fs.GetFileType(info)
		fileDBEle.Mode = fs.GetFileMode(info)
		fileDBEle.CheckSum = checkSum
		fileDBEle.LastModify = info.ModTime()

		_, exist := (*fileDB)[fpath]
		// if is the file and the path does not exist in db
		if info.IsDir() == false && exist == false {
			log.Println("client should send the file", fpath, "to server!")

			relativePath, err := filepath.Rel(baseDir, path.Dir(fpath))
			if err != nil {
				log.Fatalln("can not get relative path of the file", fpath)
			}

			err = syncing.SendFile(fpath, relativePath, deviceID)
			if err != nil {
				log.Printf("Can not send file %v: %v", fpath, err)
			}
		}

		// insert file into db
		(*fileDB)[fpath] = fileDBEle

		return nil
	}
}

// ScanDir : Scan the whole directory to update the file database
// and store some information
func ScanDir(dirPath string, deviceID string) *fs.FileDB {
	filedb := fs.GetFileDBInstance()
	filepath.Walk(dirPath, updateFileDB(filedb, deviceID, dirPath))
	return filedb
}
