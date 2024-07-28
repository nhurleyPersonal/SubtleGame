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
	conn       *websocket.Conn
	send       chan Message
	sendJSON   chan JSONMessage
	playerName string
	player     Player
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
	"guessWord": func(hub *Hub, client *Client, msg JSONMessage) {
		handleGuessWord(hub, client, msg)
	},
}

func handleJoinGame(hub *Hub, client *Client, msg JSONMessage) {
	type JoinGameBody struct {
		PlayerName string `json:"name"`
		ServerID   string `json:"serverID"`
	}

	log.Println("HUB CLIENTS 22222", hub.clients)

	var body JoinGameBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	client.playerName = body.PlayerName

	hubMultiplexer.hubs[body.ServerID].register <- client

}

func handleStartGame(hub *Hub, client *Client, msg JSONMessage) {
	body := struct {
		Leader string `json:"leader"`
	}{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}
	log.Println(body.Leader)
	if body.Leader == "leader" {
		didStart := hub.gameState.StartGame()
		if didStart {
			playersJSON, err := json.Marshal(hub.gameState.GetPlayers())
			if err != nil {
				log.Println("error marshalling players:", err)
				return
			}
			hub.broadcastJSON <- JSONMessage{
				Type: "gameStarted",
				Body: json.RawMessage(playersJSON),
			}
			log.Println("GAME STARTED")
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
	hub.broadcastJSON <- JSONMessage{
		Type: "wordSet",
		Body: json.RawMessage(fmt.Sprintf(`{"player": %s}`, strconv.Quote(client.player.Name))),
	}
}

func handleGuessWord(hub *Hub, client *Client, msg JSONMessage) {
	body := struct {
		Word     string `json:"word"`
		SelfID   string `json:"selfName"`
		TargetID string `json:"targetName"`
	}{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	completelyCorrect, partiallyCorrect, ok := hub.gameState.GuessWord(body.Word, body.SelfID, body.TargetID)
	if !ok {
		client.send <- Message{
			Type: "invalidGord",
		}
		return
	}

	if body.Word == hub.gameState.Players[body.TargetID].Word {
		client.send <- Message{
			Type: "correctWord",
		}
		hub.broadcastJSON <- JSONMessage{
			Type: "correctWord",
			Body: json.RawMessage(fmt.Sprintf(`{"player": %s}`, strconv.Quote(body.SelfID))),
		}
		return
	}

	log.Println("GUESS RESULTS", completelyCorrect, partiallyCorrect)
	client.send <- Message{
		Type: "guessResults",
		Body: fmt.Sprintf(`{"completelyCorrect": %s, "partiallyCorrect": %s}`,
			marshalList(completelyCorrect), marshalList(partiallyCorrect)),
	}
}

func marshalList(list interface{}) string {
	jsonData, err := json.Marshal(list)
	if err != nil {
		log.Println("marshal error:", err)
		return "[]"
	}
	return string(jsonData)
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
		log.Println("ReadPump: Sending client to unregister channel")
		hub.unregister <- c

		log.Println("ReadPump: Client disconnected, Read Closed")
	}()
	for {
		_, message, err := c.conn.ReadMessage() // Read a message from the WebSocket connection
		if err != nil {
			log.Println("read error:", err)
			break // Exit the loop on error
		}
		log.Println("Received JSON message:", string(message))

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

func (c *Client) WritePump(hub *Hub) {
	defer func() {
		log.Println("WritePump: Sending client to unregister channel")
		hub.unregister <- c

		log.Println("WritePump: Client disconnected, Write Closed")
	}()
	for {
		select {
		case msg := <-c.send:
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
