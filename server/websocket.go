package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/makifdb/spellcheck"
)

// Message represents a WebSocket message
type Message struct {
	Type   string `json:"type"`
	Body   string `json:"body"`
	Target string `json:"target,omitempty"`
}

// Create an upgrader with default options
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

var hubMultiplexer = NewHubMultiplexer()

// Init spellchecker
var Spellcheck, _ = spellcheck.New()

// Start the server and the hub
func Start() {
	Router()
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
