package fs

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/tonychol/sink/util"
)

// FileDBElement : The element of the db that describes the information
// of the synching directory
type FileDBElement struct {
	FileType   string
	Mode       os.FileMode
	CheckSum   string
	LastModify time.Time
}

// FileDB : The map that represents the information of each file among the whole directory
// key : each file's path string
// val : The FileDBElement struct that holds information
type FileDB map[string]FileDBElement

var instance *FileDB
var once sync.Once

// GetFileDBInstance : Using singleton to get the global filedb instance
func GetFileDBInstance() *FileDB {
	once.Do(func() {
		validConfig := make(FileDB)
		instance = &validConfig
	})
	return instance
}

func (db *FileDB) JsonStr() string {
	res, err := json.Marshal(db)
	util.HardHandleErr(err)
	return string(res[:])
}
