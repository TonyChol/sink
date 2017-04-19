package fs

import (
	"log"
	"testing"

	"github.com/tonychol/sink/config"
)

func TestSaveLinkAsJson(t *testing.T) {

}

func TestDBFileJSON(t *testing.T) {
	dbJSONPath := config.GetInstance().FileDbJSONPath
	log.Println("db file path :", dbJSONPath)
	if len(dbJSONPath) == 0 {
		t.Errorf("db file path is invalid")
	}
}
