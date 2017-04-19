package fs

import (
	"fmt"
	"testing"
)

func TestGetCheckSumOfFile(t *testing.T) {
	hash, err := getCheckSumOfFile("./db.go")
	if err == nil {
		fmt.Println(hash)
	} else {
		t.Errorf("Error occur: %s", err.Error())
	}
}
