package networking

import (
	"encoding/json"
	"net/http"
	"time"
)

// PortData contains the free port number as a part of the response
type PortData struct {
	Port int
}

// PortPayload is the resoponse when server sends the next free port to the client
type PortPayload struct {
	Status int
	Data   PortData
}

var myClient = &http.Client{Timeout: 10 * time.Second}

// GetJSON lets the client to return a JSON struct by Http GET request
func GetJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
