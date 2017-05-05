package sync

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

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

// GetAvailablePort asks the kernel for a free open port that is ready to use
func GetAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// ConnectSocket enables client to connect
func ConnectSocket(remoteAddr string) {
	log.Println("new connection: ", remoteAddr)
	connection, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Connected to server, start receiving the file name and file size")

	accpetFilesFrom(connection)
}

func accpetFilesFrom(connection net.Conn) {
	bufferSize := config.GetInstance().BufferSize
	for {
		bufferFileName := make([]byte, 64)
		bufferFileSize := make([]byte, 10)

		connection.Read(bufferFileSize)
		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		connection.Read(bufferFileName)
		fileName := strings.Trim(string(bufferFileName), ":")

		newFile, err := os.Create(fileName)

		if err != nil {
			panic(err)
		}
		defer newFile.Close()
		var receivedBytes int64

		for {
			if (fileSize - receivedBytes) < bufferSize {
				io.CopyN(newFile, connection, (fileSize - receivedBytes))
				connection.Read(make([]byte, (receivedBytes+bufferSize)-fileSize))
				break
			}
			io.CopyN(newFile, connection, bufferSize)
			receivedBytes += bufferSize
		}
		fmt.Printf("Received file %v completely!", fileName)
	}
}
