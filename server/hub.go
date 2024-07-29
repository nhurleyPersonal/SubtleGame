package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

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
		defer c.conn.Close()
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

	h.broadcast <- currentPlayersMessage
}

func (h *Hub) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case client := <-h.register:
			h.mu.Lock() // Acquire the lock
			h.clients[client] = true

			newPlayer, err := h.gameState.JoinGame(client.playerName, client, h)
			client.player = newPlayer
			if err != nil {
				log.Println("error joining game:", err)
				return
			}

			newPlayerJSON, err := json.Marshal(newPlayer)
			if err != nil {
				log.Println("error marshalling game state:", err)
			}

			playersJSON, err := json.Marshal(h.gameState.GetPlayers())
			if err != nil {
				log.Println("error marshalling players:", err)
				return
			}

			sendToPlayer := Message{
				Type: "whoami",
				Body: string(newPlayerJSON),
			}

			client.send <- sendToPlayer

			sendInitalPlayers := Message{
				Type: "currentPlayers",
				Body: string(playersJSON),
			}

			h.broadcast <- sendInitalPlayers
			h.mu.Unlock() // Release the lock

		case c := <-h.unregister:
			h.mu.Lock()
			h.CleanUp(c)
			h.mu.Unlock()

		case <-ticker.C:
			h.broadcastGameState()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
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

func (h *Hub) broadcastGameState() {
	h.mu.Lock()
	defer h.mu.Unlock()

	gameStateJSON, err := json.Marshal(h.gameState)
	if err != nil {
		log.Println("error marshalling game state:", err)
		return
	}

	gameStateMessage := Message{
		Type: "gameState",
		Body: string(gameStateJSON),
	}

	for client := range h.clients {
		select {
		case client.send <- gameStateMessage:
		default:
			close(client.send)
			delete(h.clients, client)
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
