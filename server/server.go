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
			deviceID := r.FormValue("deviceID")

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
			broadcastFile(deviceID, pool, relativePath, filename)
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
	return os.MkdirAll(baseDir, 0755)
}

func createDirIfNotExist(targetDir string) error {
	return os.MkdirAll(baseDir+targetDir, 0755)
}

func broadcastFile(deviceID string, pool networking.SocketPool, relativePath string, fname string) {
	for id, conn := range pool {
		if id != deviceID {
			sendFileInfoTo(conn, relativePath, fname)
		}
	}
}

func sendFileInfoTo(connection net.Conn, relPath string, filename string) error {

	payload := networking.FileInfoPayload{
		FileRelPath: relPath,
		FileName:    filename,
	}

	err := json.NewEncoder(connection).Encode(payload)
	if err != nil {
		return err
	}

	return nil
}
