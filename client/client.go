package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/tonychol/sink/util"
)

// postFile : Accepts a filename with relative path
// and its targetUrl of the server, posts the file to the server
func postFile(filename string, targetURL string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetURL, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(respBody))
	return nil
}

// getFilePathFromAgrs : Get the input path
// from the command-line arguments
func getFilePathFromAgrs() (string, error) {
	argsArr := os.Args
	if len(argsArr) < 2 {
		err := errors.New("You should attach a file: ./client <YOUR_FILE>")
		return "", err
	}
	return os.Args[1], nil
}

// sample usage
func main() {
	targetFile, err := getFilePathFromAgrs()
	util.HardHandleErr(err)
	targetURL := "http://localhost:8181/upload"

	postFile(targetFile, targetURL)
}
