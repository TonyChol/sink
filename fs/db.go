package fs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tonychol/sink/config"
	"github.com/tonychol/sink/util"
)

const dbJSONFileDir string = "./"
const dbJSONFileName string = "filedb.json"

// FileDBElement is the element of the db that
// describes the information of the synching directory
type FileDBElement struct {
	FileType   string
	Mode       os.FileMode
	CheckSum   string
	LastModify time.Time
	Incoming   bool
}

// NewFileDBEle returns the new FileDBElement reference
// with some default values
func NewFileDBEle() *FileDBElement {
	return &FileDBElement{Incoming: false}
}

// FileDB is a map represents the information of each file among the whole directory
// key : each file's path string
// val : The FileDBElement struct that holds information
type FileDB map[string]*FileDBElement

var instance *FileDB
var once sync.Once

// GetFileDBInstance gets the global filedb instance
func GetFileDBInstance() *FileDB {
	once.Do(func() {
		validConfig := restoreDBFromJSONFile()
		instance = &validConfig
	})
	return instance
}

// JSONStr converts the db map into the json string.
func (db *FileDB) JSONStr() string {
	res, err := json.Marshal(db)
	util.HardHandleErr(err)
	return string(res[:])
}

// SaveDBAsJSON persists the db map instance into json file.
func (db *FileDB) SaveDBAsJSON() {
	absPath, err := filepath.Abs(config.GetInstance().FileDbJSONPath)
	util.PanicIf(err)

	log.Println("db persisted in", config.GetInstance().FileDbJSONPath)

	os.Remove(absPath)

	b, err := json.Marshal(db)
	util.PanicIf(err)

	err = ioutil.WriteFile(absPath, b, 0666)
	util.PanicIf(err)
}

// AddFileDir accepts a file or dir string
// and creates a FileDBEle based on that.
func (db *FileDB) AddFileDir(fpath string) {
	ele := NewFileDBEle()
	(*db)[fpath] = ele
	db.SaveDBAsJSON()
}

// AddIncomingFileDir accepts a file or dir string
// and creates a FileDBEle based on that.
// Note that the incoming attribute needs to be true.
func (db *FileDB) AddIncomingFileDir(fpath string) {
	db.AddFileDir(fpath)
	(*db)[fpath].Incoming = true
}

// UnsetIncoming will set the Incoming attribute
// into false. It happens when the incoming file
// has been created and dumpped into the file system.
func (db *FileDB) UnsetIncoming(fpath string) {
	log.Println("unsetting incoming", fpath)
	(*db)[fpath].Incoming = false
	db.SaveDBAsJSON()
}

// restoreDBFromJSONFile : Try to restore the db instance from the json file
// if this file does not exists, then return an empty FileDB instance.
// Note that this function will only be called by `GetFileDBInstance()`
func restoreDBFromJSONFile() FileDB {
	absPath, err := filepath.Abs(config.GetInstance().FileDbJSONPath)
	util.PanicIf(err)
	dat, err := ioutil.ReadFile(absPath)
	if err != nil {
		return make(FileDB)
	}

	var dbInstance FileDB
	err = json.Unmarshal(dat, &dbInstance)
	util.PanicIf(err)

	return dbInstance
}
