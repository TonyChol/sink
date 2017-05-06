package networking

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

// SocketPool saves the references pointing to all the connecting connection instances
type SocketPool map[string]net.Conn

// ServeWithSocket lets the server start a new port that serves one client's connection
func ServeWithSocket(newport int, pool *SocketPool) {
	addrWithPort := fmt.Sprintf("localhost:%v", newport)
	server, err := net.Listen("tcp", addrWithPort)
	defer server.Close()
	if err != nil {
		fmt.Printf("ServeWithSocket: Error setting up socket with port: %v: %v", addrWithPort, err)
		os.Exit(1)
	}

	fmt.Println("ServeWithSocket: Socket server started! Waiting for connections...")
	for {
		// Wait for a connection.
		connection, err := server.Accept()
		if err != nil {
			log.Println("ServeWithSocket: Accept socket request from client error: ", err)
			os.Exit(1)
		}

		// receive the message
		var deviceID string
		err = gob.NewDecoder(connection).Decode(&deviceID)
		if err != nil {
			log.Fatalln("Can not get device id from the client via socket")
		}

		// Put the connection inside the pool
		(*pool)[deviceID] = connection

		defer func() {
			delete(*pool, deviceID)
			connection.Close()
		}()

		for id, conn := range *pool {
			log.Printf("Connection %v: %v --> %v", id, conn.RemoteAddr().String(), conn.LocalAddr().String())
		}

		fmt.Printf("ServeWithSocket: Client %v connected from %v\n", deviceID, connection.RemoteAddr())
	}
}
