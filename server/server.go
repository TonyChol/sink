package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tonychol/sink/config"
	"github.com/tonychol/sink/networking"
	"github.com/tonychol/sink/sync"
	"github.com/tonychol/sink/util"
)

const baseDir = "." + string(filepath.Separator) + "sync" + string(filepath.Separator)

func main() {
	pool := make(networking.SocketPool)
	err := createBaseDir()
	if err != nil {
		log.Println("can not create baseDir: ", err)
	}

	http.HandleFunc("/upload", upload(pool))
	http.HandleFunc("/socketPort", getFreePort(pool))
	http.Handle("/", http.FileServer(http.Dir("sync")))
	sevrAddr := fmt.Sprintf(":%d", config.GetInstance().DevPort)
	log.Printf("Server has been set up at :%v\n ", sevrAddr)
	err = http.ListenAndServe(sevrAddr, nil) // set listen port
	util.HardHandleErr(err)
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}

		log.Println("Message received:", string(buf[0:]))
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

// upload logic
func upload(pool networking.SocketPool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			log.Println("method:", r.Method)
			crutime := time.Now().Unix()
			h := md5.New()
			io.WriteString(h, strconv.FormatInt(crutime, 10))
			token := fmt.Sprintf("%x", h.Sum(nil))

			t, _ := template.ParseFiles("upload.gtpl")
			t.Execute(w, token)
		} else if r.Method == "POST" {
			log.Println("Getting post request from ", r.RemoteAddr)
			relativePath := r.FormValue("relativePath")
			filename := r.FormValue("filename")

			r.ParseMultipartForm(32 << 20)
			file, handler, err := r.FormFile("uploadfile")
			defer file.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Fprintf(w, "%v", handler.Header)

			// Try to create the directory to hold the target incoming file if not exist
			err = createDirIfNotExist(relativePath)
			util.HardHandleErr(err)

			targetFilePath := baseDir + relativePath + string(filepath.Separator) + filename
			f, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				return
			}
			log.Println("Method:", r.Method, ", file", filename, "has been received from client")
			defer f.Close()
			io.Copy(f, file)

			// start to broadcast the file
			broadcastFile(r.RemoteAddr, pool, relativePath, filename)
		}
	}
}

// getFreePort is a handler function that accept an Http GET request and
// send the next free available port back to the client.
func getFreePort(pool networking.SocketPool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {

			var res networking.PortPayload
			newport, err := sync.GetAvailablePort()
			// Start new port serving a new socket connection
			go networking.ServeWithSocket(newport, &pool)

			// Making response
			if err != nil {
				log.Println("can not get a free port in server, ", err)
				res = networking.PortPayload{
					Status: http.StatusServiceUnavailable,
					Data: networking.PortData{
						Port: 0,
					},
				}
			} else {
				res = networking.PortPayload{
					Status: http.StatusOK,
					Data: networking.PortData{
						Port: newport,
					},
				}
			}

			// send response as json
			json.NewEncoder(w).Encode(res)

		} else {
			resErr := networking.PortPayload{
				Status: http.StatusMethodNotAllowed,
				Data: networking.PortData{
					Port: 0,
				},
			}
			json.NewEncoder(w).Encode(resErr)
		}
	}
}

func createBaseDir() error {
	return os.MkdirAll(baseDir, 0777)
}

func createDirIfNotExist(targetDir string) error {
	return os.MkdirAll(baseDir+targetDir, 0777)
}

func sendFileToClient(connection net.Conn, relativePath string, fname string) {
	file, err := os.Open(baseDir + relativePath + string(filepath.Separator) + fname)
	if err != nil {
		log.Fatal("sendFileToClient: can not open file", fname)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal("sendFileToClient: can not get file info of", fname)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fileRelPath := fillString(relativePath, 64)
	fmt.Println("sendFileToClient: Sending filename, filesize and relative path of", fname)
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	connection.Write([]byte(fileRelPath))
	sendBuffer := make([]byte, config.GetInstance().BufferSize)
	fmt.Println("sendFileToClient: Start sending binary of file", fname)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("sendFileToClient: File has been sent, closing connection!")
	return
}

func fillString(returnString string, toLength int) string {
	for {
		strSize := len(returnString)
		if strSize < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}

func broadcastFile(sourceAddr string, pool networking.SocketPool, relativePath string, fname string) {
	for conn := range pool {
		log.Println("broadcastFileExcept: source addr = ", sourceAddr)
		log.Println("broadcastFileExcept: connection's remote = ", conn.RemoteAddr().String())
		if sourceAddr != conn.RemoteAddr().String() {
			sendFileToClient(conn, relativePath, fname)
		}
	}
}
