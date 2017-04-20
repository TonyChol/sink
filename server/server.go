/* ThreadedIPEchoServer
 */
package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"path/filepath"

	"github.com/tonychol/sink/util"
)

const baseDir = "." + string(filepath.Separator) + "sync" + string(filepath.Separator)

func main() {
	http.HandleFunc("/upload", upload)
	log.Println("Server has been set up at :8181")
	err := http.ListenAndServe(":8181", nil) // set listen port
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
func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		log.Println("method:", r.Method)
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		relativePath := r.FormValue("relativePath")
		filename := r.FormValue("filename")

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
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
		log.Println("Method:", r.Method, " , file", filename, " has been received from client")
		defer f.Close()
		io.Copy(f, file)
	}
}

func createDirIfNotExist(targetDir string) error {
	return os.MkdirAll(baseDir+targetDir, 0777)
}
