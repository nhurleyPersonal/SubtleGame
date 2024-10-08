package server

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func connectionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HERE 4", r.URL)
	lobbyID := strings.ToUpper(r.URL.Query().Get("lobbyID"))
	playerName := r.URL.Query().Get("name")
	playerID := r.URL.Query().Get("id")
	log.Println(lobbyID, playerName, playerID)

	// Check if the request is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(r) {
		err := clientJoinsLobby(w, r, lobbyID, playerName, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func clientJoinsLobby(w http.ResponseWriter, r *http.Request, lobbyID string, playerName string, playerID string) error {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("connection error:", err)
		return errors.New("connection error")
	}

	hub, ok := hubMultiplexer.hubs[lobbyID]
	if !ok {
		log.Println("Hub not found")
		return errors.New("hub not found")
	}

	client := &Client{conn: ws, send: make(chan Message, 256), sendJSON: make(chan JSONMessage, 256), sendhtml: make(chan string, 256)}
	client.playerName = playerName
	client.playerID = playerID

	// Reconnecting logic
	player, ok := hub.gameState.Players[playerID]
	if ok {
		log.Println("FOUND PLAYER", player.Name)
		client.player = player
		hub.reconnect <- client
	} else {
		log.Println("REGISTERING PLAYER", client.playerName)
		hub.register <- client
	}

	go client.ReadPump(hub)
	go client.WritePump(hub)

	return nil
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "server/static/favicon.ico")
}

func gameLobbyHandler(w http.ResponseWriter, r *http.Request) {
	lobbyID := strings.ToUpper(r.URL.Query().Get("lobbyID"))
	playerName := r.URL.Query().Get("name")
	playerID := uuid.New().String()

	hub, ok := hubMultiplexer.hubs[lobbyID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Could not find requested lobby!"))
		return
	}

	if hub.gameState.Started {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("This game has already started."))
		return
	}

	for _, player := range hub.gameState.Players {
		if player.Name == playerName {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("Name taken already."))
			return
		}
	}

	SendIntialStartGame(r, w, lobbyID, playerName, playerID)
}

func createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	hub, lobbyID := hubMultiplexer.CreateHub()
	hub.gameState = NewGameState()
	go hub.run()
	htmlContent := `<div class="server-created-text">` + lobbyID + `</div>`
	w.Write([]byte(htmlContent))
}

func Router() {
	http.HandleFunc("/ws", connectionHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/createServer", createLobbyHandler)
	http.HandleFunc("/gamelobby", gameLobbyHandler)
	http.HandleFunc("/youmayenter", connectionHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/static/index.html", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}
