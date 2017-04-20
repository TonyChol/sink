package fs

import (
	"fmt"
	"testing"
)

func TestGetCheckSumOfFile(t *testing.T) {
	hash, err := GetCheckSumOfFile("./db.go")
	if err == nil {
		fmt.Println(hash)
	} else {
		t.Errorf("Error occur: %s", err.Error())
	}
}

func TestGetRelativeDirFromRoot(t *testing.T) {
	tarpath := "/Users/tonychol/go/src/github.com/tonychol/sink/testDir/subDir/two.txt"
	expected := "subDir"
	res, err := GetRelativeDirFromRoot(tarpath)
	if err != nil {
		t.Errorf("Error occur: %s", err.Error())
	}

	if res != expected {
		t.Errorf("Relative path is wrong. Expected %s but get %s", expected, res)
	}
}

func TestGetFileNameFromFilePath(t *testing.T) {
	fpath := "/Users/tonychol/go/src/github.com/tonychol/sink/fs/db_test.go"
	res := GetFileNameFromFilePath(fpath)
	if res != "db_test.go" {
		t.Errorf("File Name is wrong")
	}

	anotherFilePath := "/Users/tonychol/go/src/github.com/tonychol/sink/testDir/subDir/two.txt"
	expected := "two.txt"
	if GetFileNameFromFilePath(anotherFilePath) != expected {
		t.Errorf("File Name is wrong. Expected %s but get %s", expected, res)
	}
}
