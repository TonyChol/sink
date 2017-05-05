package networking

import (
	"fmt"
	"net"
	"os"
)

// SocketPool saves the references pointing to all the connecting connection instances
type SocketPool map[net.Conn]bool

// ServeWithSocket lets the server start a new port that serves one client's connection
func ServeWithSocket(newport int, pool *SocketPool) {
	addrWithPort := fmt.Sprintf("localhost:%v", newport)
	server, err := net.Listen("tcp", addrWithPort)
	defer server.Close()
	if err != nil {
		fmt.Printf("Error setting up socket with port: %v: %v", addrWithPort, err)
		os.Exit(1)
	}

	fmt.Println("Socket server started! Waiting for connections...")
	for {
		// Wait for a connection.
		connection, err := server.Accept()
		defer func() {
			delete(*pool, connection)
			connection.Close()
		}()
		if err != nil {
			fmt.Println("Accept socket request from client error: ", err)
			os.Exit(1)
		}

		// Put the connection inside the pool
		(*pool)[connection] = true

		fmt.Printf("Client %v connected", connection.RemoteAddr())
	}
}
