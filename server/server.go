/* ThreadedIPEchoServer
 */
package main

import (
	"fmt"
	"net"
	"os"

	"log"

	"github.com/tonychol/sink/util"
)

func main() {
	service := "127.0.0.1:8181"
	listener, err := net.Listen("tcp", service)
	util.HardHandleErr(err)

	log.Println("Server has been set up at service", service)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
