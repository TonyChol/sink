package networking

// PortData contains the free port number as a part of the response
type PortData struct {
	Port int
}

// PortPayload is the resoponse when server sends the next free port to the client
type PortPayload struct {
	Status int
	Data   PortData
}
