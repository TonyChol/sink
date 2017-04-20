package sync

import (
	"testing"
)

func TestGetTargetURLFromConfig(t *testing.T) {
	targetURL := getTargetURLFromConfig()
	if targetURL != "http://localhost:8181/upload" {
		t.Errorf("Expected http://localhost:8181/upload but get %s", targetURL)
	}
}

func TestSendFile(t *testing.T) {
	fileToSend := "./sync.go"
	SendFile(fileToSend)
}
