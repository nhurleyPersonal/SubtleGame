package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients        map[*Client]bool
	broadcast      chan Message
	broadcastJSON  chan JSONMessage
	register       chan *Client
	unregister     chan *Client
	gameState      GameState
	connToPlayerId map[*websocket.Conn]string
	mu             sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan Message),
		broadcastJSON:  make(chan JSONMessage),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		gameState:      GameState{},
		connToPlayerId: make(map[*websocket.Conn]string),
	}
}

func (h *Hub) GetGameState() *GameState {
	return &h.gameState
}

func (h *Hub) CleanUp(c *Client) {
	log.Println("Cleaning up client")
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
		close(c.sendJSON)
		h.gameState.RemovePlayer(c.player)
		c.conn.Close()
	}

	currentPlayers, err := json.Marshal(h.gameState.GetPlayers())
	if err != nil {
		log.Println("error marshalling game state:", err)
		return
	}

	currentPlayersMessage := Message{
		Type: "currentPlayers",
		Body: string(currentPlayers),
	}

	log.Println("Broadcasting currentPlayersMessage:", currentPlayersMessage.Body, currentPlayersMessage.Type)
	h.broadcast <- currentPlayersMessage
	log.Println("Sent current players message", currentPlayersMessage)
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock() // Acquire the lock

			h.clients[client] = true

			currentPlayers, err := json.Marshal(h.gameState.GetPlayers())
			if err != nil {
				log.Println("error marshalling game state:", err)
				return
			}

			sendInitalPlayers := Message{
				Type: "currentPlayers",
				Body: string(currentPlayers),
			}

			client.send <- sendInitalPlayers
			h.mu.Unlock() // Release the lock

		case c := <-h.unregister:
			log.Println("Unregistering client", c)
			h.mu.Lock() // Acquire the lock
			defer h.mu.Unlock()
			h.CleanUp(c)

		case message := <-h.broadcast:
			log.Println("BROADCASTING MESSAGE", message)
			h.mu.Lock() // Acquire the lock
			for client := range h.clients {
				log.Println("Starting broadcast")
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock() // Release the lock

		case message := <-h.broadcastJSON:
			h.mu.Lock() // Acquire the lock
			for client := range h.clients {
				select {
				case client.sendJSON <- message:
				default:
					close(client.sendJSON)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock() // Release the lock
		}
	}
}

type HubMultiplexer struct {
	hubs map[string]*Hub
	mu   sync.Mutex
}

func NewHubMultiplexer() *HubMultiplexer {
	return &HubMultiplexer{
		hubs: make(map[string]*Hub),
	}
}

func (hm *HubMultiplexer) CreateHub() (*Hub, string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	lobbyID := generateRandomString(6)
	hub := NewHub()
	hm.hubs[lobbyID] = hub
	go hub.run()
	return hub, lobbyID
}

func (hm *HubMultiplexer) GetHub(lobbyID string) *Hub {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hub, exists := hm.hubs[lobbyID]
	if !exists {
		log.Println("Hub not found")
		return nil
	}
	return hub
}

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
