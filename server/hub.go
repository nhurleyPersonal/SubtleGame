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
	reconnect      chan *Client
	unregister     chan *Client
	gameState      GameState
	connToPlayerId map[*websocket.Conn]string
	serverCode     string
	mu             sync.Mutex
}

func NewHub(lobbyID string) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan Message),
		broadcastJSON:  make(chan JSONMessage),
		register:       make(chan *Client),
		reconnect:      make(chan *Client),
		unregister:     make(chan *Client),
		gameState:      GameState{},
		serverCode:     lobbyID,
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
		c.conn.Close()
	}

	log.Println("Attempting Reconnection")
	// Handle reconnection
	func() {
		timeout := time.After(1 * time.Minute)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				log.Println("Reconnection attempts timed out")
				h.mu.Lock()
				h.gameState.RemovePlayer(c.player)
				h.mu.Unlock()
				return
			case <-ticker.C:

				// // Construct the WebSocket URL
				// url := "wss://" + c.conn.RemoteAddr().String()
				// log.Println("Attempting to reconnect to:", url)

				// // Print headers if needed (example, adjust as necessary)
				// headers := make(http.Header)
				// headers.Add("Origin", "http://example.com")
				// log.Println("Request headers:", headers)

				// newConn, _, err := websocket.DefaultDialer.Dial(c.conn.RemoteAddr().String(), nil)
				// if err != nil {
				// 	log.Println("Reconnection attempt failed:", err)
				// 	continue
				// }
				// c.conn = newConn
				// h.register <- c
				return
			}
		}
	}()

	currentPlayers, err := json.Marshal(h.gameState.GetPlayers())
	if err != nil {
		log.Println("error marshalling game state:", err)
		return
	}

	currentPlayersMessage := Message{
		Type: "currentPlayers",
		Body: string(currentPlayers),
	}

	h.broadcastMessage(currentPlayersMessage)
}

func (hm *HubMultiplexer) shutdownHub(hubCode string) {
	hubToBeCleaned := hm.hubs[hubCode]
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hubToBeCleaned.mu.Lock()
	defer hubToBeCleaned.mu.Unlock()
	if len(hubToBeCleaned.clients) != 0 {
		return
	}
	for client := range hubToBeCleaned.clients {
		client.conn.Close()
		delete(hubToBeCleaned.clients, client)
	}
	log.Println("Shut down hub", hubCode)
}

func (h *Hub) run() {
	ticker := time.NewTicker(10 * time.Second)
	shutdownTimer := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	defer shutdownTimer.Stop()
	for {
		select {
		case client := <-h.register:
			h.mu.Lock() // Acquire the lock
			shutdownTimer.Stop()
			h.clients[client] = true

			newPlayer, err := h.gameState.JoinGame(client.playerName, client, h)
			client.player = newPlayer
			if err != nil {
				log.Println("error joining game:", err)
				return
			}

			for c := range h.clients {
				log.Println("CLIENTS IN HUB", c.player.Name)
			}

			for p := range h.gameState.Players {
				log.Println("PLAYERS IN GAME", h.gameState.Players[p].Name)
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

			log.Println("SENT TO PLAYER")

			sendInitalPlayers := Message{
				Type: "currentPlayers",
				Body: string(playersJSON),
			}

			log.Println("SENT TO PLAYER 2")

			h.broadcastMessage(sendInitalPlayers)

			log.Println("SENT TO ALL PLAYERS")
			h.mu.Unlock() // Release the lock

		case c := <-h.reconnect:
			h.mu.Lock()
			shutdownTimer.Stop()
			h.clients[c] = true
			log.Println("CLIENT RECONNECT DEETS", c.playerID, c.playerName)

			reconnectPlayer, err := h.gameState.ReconnectPlayer(c, h)
			if err != nil {
				log.Println("error joining game:", err)
				continue
			}
			c.player = reconnectPlayer

			log.Println("RECONNECTED PLAYER", c.player.Name)

			reconnecPlayerJSON, err := json.Marshal(reconnectPlayer)
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
				Body: string(reconnecPlayerJSON),
			}

			c.send <- sendToPlayer

			sendInitalPlayers := Message{
				Type: "currentPlayers",
				Body: string(playersJSON),
			}

			h.broadcastMessage(sendInitalPlayers)
			h.mu.Unlock() // Release the lock

		case c := <-h.unregister:
			h.mu.Lock()
			shutdownTimer.Reset(20 * time.Second)
			h.CleanUp(c)
			h.mu.Unlock()

		case <-ticker.C:
			h.broadcastGameState()

		case <-shutdownTimer.C:
			hubMultiplexer.shutdownHub(h.serverCode)
			return

		case message := <-h.broadcast:
			h.mu.Lock()
			log.Println("BROADCASTING TO ALL PLAYERS", message)
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

func (h *Hub) broadcastMessage(message Message) {
	h.mu.Lock()
	log.Println("BROADCASTING TO ALL PLAYERS", message)
	for client := range h.clients {
		client.send <- message
	}
	h.mu.Unlock() // Release the lock

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
		client.send <- gameStateMessage
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
	hub := NewHub(lobbyID)
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
