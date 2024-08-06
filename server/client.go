package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"

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
	sendhtml   chan string
	sendJSON   chan JSONMessage
	playerName string
	playerID   string
	player     Player
}

type JSONMessage struct {
	Type    string          `json:"type"`
	Body    json.RawMessage `json:"body"`
	Headers json.RawMessage `json:"HEADERS"`
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

	var body JoinGameBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	client.playerName = body.PlayerName

	hubMultiplexer.hubs[body.ServerID].register <- client

}

func handleStartGame(hub *Hub, client *Client, msg JSONMessage) {

	playerList := make([]Player, 0)

	for c, _ := range hub.clients {
		log.Println("CLIENT", c)
		playerList = append(playerList, c.player)
	}

	log.Println(playerList)

	for c := range hub.clients {
		clientPlayer := c.player

		// Execute the self player template
		gameroomTmpl, err := template.ParseFiles("server/templates/gameStartedRoom.html")
		if err != nil {
			log.Println("template parse error:", err)
			return
		}

		var tpl bytes.Buffer
		tplData := struct {
			SelfPlayerName  string
			SelfPlayerID    string
			SelfPlayerScore int
			Word            string
			OtherPlayers    []Player
		}{
			SelfPlayerName:  clientPlayer.Name,
			SelfPlayerID:    clientPlayer.ID,
			SelfPlayerScore: clientPlayer.Score,
			Word:            clientPlayer.Word,
			OtherPlayers:    playerList,
		}

		if err := gameroomTmpl.Execute(&tpl, tplData); err != nil {
			log.Println("template execute error:", err)
			return
		}

		c.sendhtml <- tpl.String()
	}

}

func handleSetWord(hub *Hub, client *Client, msg JSONMessage) {
	word := ""

	if err := json.Unmarshal(msg.Body, &word); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	err := hub.gameState.SetWord(&client.player, word)
	if err != nil {
		client.send <- Message{
			Type: "invalidWord",
		}
		return
	}

	everyoneReady := true

	for _, pReady := range hub.gameState.Players {
		log.Println("ISREADY", pReady.Name, pReady.Ready)
		if !pReady.Ready {
			everyoneReady = false
		}
	}

	if everyoneReady {
		UpdateStartButton(hub)
	}

	// Execute the self player template
	tmpl, err := template.ParseFiles("server/templates/readyPlayerName.html")
	if err != nil {
		log.Println("template parse error:", err)
		return
	}

	var tpl bytes.Buffer
	data := struct {
		PlayerName string
	}{
		PlayerName: client.player.Name,
	}

	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Println("template execute error:", err)
		return
	}

	// Execute the broadcast to all players
	readyBroadcastTmpl, err := template.ParseFiles("server/templates/playerReadyLobby.html")
	if err != nil {
		log.Println("template parse error:", err)
		return
	}

	var tplBroadcast bytes.Buffer
	dataBroadcast := struct {
		PlayerName string
	}{
		PlayerName: client.player.Name,
	}

	if err := readyBroadcastTmpl.Execute(&tplBroadcast, dataBroadcast); err != nil {
		log.Println("template execute error:", err)
		return
	}

	client.sendhtml <- tpl.String()
	hub.broadcasthtml <- tplBroadcast.String()

}

type LetterAndClass struct {
	Letter string
	Class  string
}

func handleGuessWord(hub *Hub, client *Client, msg JSONMessage) {
	var body []string
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	if len(body) != 2 {
		log.Println("invalid body length")
		return
	}

	guess := body[0]
	targetPlayer := strings.TrimPrefix(body[1], "id-")

	lettersMapped := make([]LetterAndClass, 0, len(guess))

	completelyCorrect, partiallyCorrect, ok := hub.gameState.GuessWord(guess, client.player.ID, targetPlayer)
	if !ok {
		client.send <- Message{
			Type: "invalidGuess",
		}
		return
	}

	guessingPlayer := hub.gameState.Players[client.player.ID]
	// Execute correct guess template if its a correct guess
	if len(completelyCorrect) == len(guess) {
		BroadcastScoreUpdate(hub, guessingPlayer.ID, guessingPlayer.Score)
		targetPlayerObj := hub.gameState.Players[targetPlayer]
		ok = SendCorrectGuess(client, targetPlayerObj, guess)
		if !ok {
			client.send <- Message{
				Type: "invalidGuess",
			}
			return
		}
		return
	}

	// if not completely correct map the correct/partially correct letters to template
	for i := 0; i < len(guess); i++ {
		letter := string(guess[i])
		class := ""

		if len(completelyCorrect) > 0 {
			for _, num := range completelyCorrect {
				if i == num {
					class = "letter-correct"
				}
			}
		}

		if len(partiallyCorrect) > 0 {
			for _, num := range partiallyCorrect {
				if i == num {
					class = "letter-partially-correct"
				}
			}
		}

		lettersMapped = append(lettersMapped, LetterAndClass{letter, class})
	}

	ok = SendGuessResults(client, targetPlayer, lettersMapped)
	if !ok {
		client.send <- Message{
			Type: "invalidGuess",
		}
		return
	}

	// Reset the input container
	ok = SendCardReset(client, targetPlayer)
	if !ok {
		client.send <- Message{
			Type: "invalidGuess",
		}
		return
	}

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
		hub.unregister <- c
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
		msgStr := string(message)
		log.Println(msgStr) // Changed to log.Println
		if handler, found := messageHandlers[msgJSON.Type]; found {
			handler(hub, c, msgJSON)
		} else {

			log.Printf("Unhandled message type: %s", msg.Type)
		}
	}
}

func (c *Client) WritePump(hub *Hub) {
	defer func() {
		hub.unregister <- c
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
		case msg := <-c.sendhtml:
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}
}
