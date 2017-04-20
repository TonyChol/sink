package sync

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/tonychol/sink/config"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/util"
)

// SendFile : Public api that accept a file name(path) and send it to the server
func SendFile(filename string) error {
	targetURL := getTargetURLFromConfig()
	return postFile(filename, targetURL)
}

// getTargetURLFromConfig : get the target url by parsing the config file
func getTargetURLFromConfig() string {
	conf := config.GetInstance()
	targetURL := conf.DevServer + fmt.Sprintf(":%d", conf.DevPort) + conf.DevUploadURLPattern
	return targetURL
}

// postFile : Accepts a filename which is a relative path
// and its targetUrl of the server, posts the file to the server
func postFile(filename string, targetURL string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	relativePath, err := fs.GetRelativeDirFromRoot(filename)
	util.HardHandleErr(err)
	fileFullName := fs.GetFileNameFromFilePath(filename)
	log.Println("relative path =", relativePath)
	log.Println("file full name =", fileFullName)
	// this step is very important
	bodyWriter.WriteField("relativePath", relativePath)
	bodyWriter.WriteField("filename", fileFullName)
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

// // sample usage
// func main() {
// 	targetFile, err := getFilePathFromAgrs()
// 	util.HardHandleErr(err)
// 	targetURL := "http://localhost:8181/upload"
// 	postFile(targetFile, targetURL)
// }
