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
type FileDB map[string]FileDBElement

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

// JSONStr converts the db map into the json string
func (db *FileDB) JSONStr() string {
	res, err := json.Marshal(db)
	util.HardHandleErr(err)
	return string(res[:])
}

// SaveDBAsJSON :Persist the db map instance into json file
func (db *FileDB) SaveDBAsJSON() {
	absPath, err := filepath.Abs(config.GetInstance().FileDbJSONPath)
	util.PanicIf(err)

	log.Println("db persistance: db json path =", absPath)

	os.Remove(absPath)

	b, err := json.Marshal(db)
	util.PanicIf(err)

	err = ioutil.WriteFile(absPath, b, 0666)
	util.PanicIf(err)
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
