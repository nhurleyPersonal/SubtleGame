package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

type SendPlayerInfoMessage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Client represents a WebSocket client
type Client struct {
	conn     *websocket.Conn
	send     chan Message
	sendJSON chan JSONMessage
	player   Player
}

type JSONMessage struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

var messageHandlers = map[string]func(*Hub, *Client, JSONMessage){
	"joinGame": func(hub *Hub, client *Client, msg JSONMessage) {
		handleJoinGame(hub, client, msg)
	},
	"showPage": func(hub *Hub, client *Client, msg JSONMessage) {
		handleShowPage(client, msg)
	},
	"startGame": func(hub *Hub, client *Client, msg JSONMessage) {
		handleStartGame(hub, client, msg)
	},
	"setWord": func(hub *Hub, client *Client, msg JSONMessage) {
		handleSetWord(hub, client, msg)
	},
}

func handleJoinGame(hub *Hub, client *Client, msg JSONMessage) {
	type JoinGameBody struct {
		PlayerName string `json:"name"`
		ServerID   string `json:"serverID"`
	}

	var body JoinGameBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	hub, ok := hubMultiplexer.hubs[body.ServerID]
	if !ok {
		log.Println("Hub not found")
		return
	}
	hub.register <- client
	newPlayer, err := hub.gameState.JoinGame(body.PlayerName, client, hub)
	if err != nil {
		log.Println("error joining game:", err)
		return
	}

	newPlayerJSON, err := json.Marshal(newPlayer)
	if err != nil {
		log.Println("error marshalling game state:", err)
	}

	sendToEveryone := Message{
		Type: "playerJoined",
		Body: string(newPlayerJSON),
	}

	sendToPlayer := Message{
		Type: "whoami",
		Body: string(newPlayerJSON),
	}

	hub.broadcast <- sendToEveryone
	client.send <- sendToPlayer
}

func handleStartGame(hub *Hub, client *Client, msg JSONMessage) {
	body := struct {
		Leader string `json:"leader"`
	}{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}
	if body.Leader == "leader" {
		hub.gameState.StartGame()
		hub.broadcast <- Message{
			Type: "gameStarted",
		}
	} else {
		client.send <- Message{
			Type: "invalidLeader",
		}
	}

}

func handleSetWord(hub *Hub, client *Client, msg JSONMessage) {
	body := struct {
		Word string `json:"word"`
	}{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}
	err := hub.gameState.SetWord(&client.player, body.Word)
	if err != nil {
		client.send <- Message{
			Type: "invalidWord",
		}
		return
	}
}

func handleIndexPage(client *Client, msg Message) {
	// Send a message back to the client to navigate to the game lobby
	response := JSONMessage{
		Type: "navigate",
		Body: json.RawMessage(`{"url": "/index.html"}`),
	}
	client.sendJSON <- response
}

func handleShowPage(client *Client, msg JSONMessage) {
	var url string
	if err := json.Unmarshal(msg.Body, &url); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	htmlContent, err := os.ReadFile("static/" + url + ".html")
	if err != nil {
		log.Println("file read error:", err)
		return
	}

	response := JSONMessage{
		Type: "navigate",
		Body: json.RawMessage(fmt.Sprintf(`{"html": %s}`, strconv.Quote(string(htmlContent)))),
	}

	client.sendJSON <- response
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		if c.player.ID != "" {
			log.Println("Closing WebSocket connection in ReadPump")
			log.Println("Closing websocket for client", c.player.ID) // Debugging log
			hub.unregister <- c
		} else {
			c.conn.Close()
		}
	}()
	for {
		_, message, err := c.conn.ReadMessage() // Read a message from the WebSocket connection
		if err != nil {
			log.Println("read error:", err)
			break // Exit the loop on error
		}

		var msg Message
		var msgJSON JSONMessage
		if err := json.Unmarshal(message, &msgJSON); err != nil { // Unmarshal the JSON message
			log.Println("unmarshal error:", err)
			continue // Continue to the next iteration on error
		}
		if handler, found := messageHandlers[msgJSON.Type]; found {
			handler(hub, c, msgJSON)
		} else {
			log.Printf("Unhandled message type: %s", msg.Type)
		}
	}
}

func (c *Client) WritePump() {

	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			log.Println("Sending message from send channel:", msg) // Debugging log
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("write error:", err)
				return
			}
		case msg := <-c.sendJSON:
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}
}
