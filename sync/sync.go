package sync

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"

	"github.com/tonychol/sink/config"
	"github.com/tonychol/sink/fs"
	"github.com/tonychol/sink/networking"
)

// SendFile : Public api that accept a file name(path) and send it to the server
func SendFile(filename string, relativePath string, deviceID string) error {
	targetURL := getTargetURLFromConfig()
	return postFileToServer(filename, relativePath, targetURL, deviceID)
}

// getTargetURLFromConfig : get the target url by parsing the config file
func getTargetURLFromConfig() string {
	conf := config.GetInstance()
	targetURL := "http://" + conf.DevServer + fmt.Sprintf(":%d", conf.DevPort) + conf.DevUploadURLPattern
	return targetURL
}

// postFileToServer accepts a filename which is a relative path
// and its targetUrl of the server, posts the file to the server,
// finally it attaches the deviceID so that the server could
// distinguish the client
func postFileToServer(filename string, relativePath string, targetURL string, deviceID string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileFullName := fs.GetFileNameFromFilePath(filename)
	log.Printf("post file to server, relative path: %v, name: %v", relativePath, fileFullName)
	// this step is very important
	bodyWriter.WriteField("relativePath", relativePath)
	bodyWriter.WriteField("filename", fileFullName)
	bodyWriter.WriteField("deviceID", deviceID)
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
func ConnectSocket(remoteAddr string, targetDir string, deviceID string) {
	log.Println("ConnectSocket: new connection: ", remoteAddr)

	connection, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// send device id to server
	err = gob.NewEncoder(connection).Encode(deviceID)
	if err != nil {
		log.Println("can not send devide id to server", err)
	}

	fmt.Println("Client connected to server, start receiving the file name and its relative path")

	waitRemoteFileInfo(connection, targetDir)
}

func waitRemoteFileInfo(connection net.Conn, targetFullDir string) {
	for {
		var fileinfo networking.FileInfoPayload
		// Get available port for socket connection from server
		json.NewDecoder(connection).Decode(&fileinfo)

		log.Println("Get file info from server", fileinfo.FileRelPath, fileinfo.FileName)

		// Enroll incoming files
		go getFile(fileinfo.FileRelPath, fileinfo.FileName, targetFullDir)
		// Exit incoming files
	}
}

func getFile(relPath, filename, targetFullDir string) {
	serverAddr := "http://" + config.GetInstance().DevServer + fmt.Sprintf(":%d", config.GetInstance().DevPort)
	endpoint := serverAddr + "/" + relPath + "/" + filename

	err := os.MkdirAll(targetFullDir+string(os.PathSeparator)+relPath, 0755)
	if err != nil {
		log.Fatalf("can not create folder '%v' for incoming file\n", relPath)
	}

	fileFullPath := targetFullDir + string(os.PathSeparator) + relPath + string(os.PathSeparator) + filename

	// insert the file into filedb, with the incoming attribute set to true
	db := fs.GetFileDBInstance()
	if _, exists := (*db)[fileFullPath]; exists == true {
		a := (*db)[fileFullPath]
		(*a).Incoming = true
	} else {
		db.AddIncomingFileDir(fileFullPath)
	}

	out, err := os.Create(fileFullPath)
	if err != nil {
		log.Fatalf("can not create file %v\n", fileFullPath)
	}
	defer out.Close()

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatalf("can not get file from endpoint %v\n", endpoint)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalln("can not download file from %v", endpoint)
	}
}
