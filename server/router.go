package server

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func connectionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	serverID := r.URL.Query().Get("serverID")
	playerID := r.URL.Query().Get("playerID")
	hub, ok := hubMultiplexer.hubs[serverID]
	if !ok {
		log.Println("Hub not found")
		return
	}

	// Check if the request is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(r) {
		err := clientJoinsLobby(w, r, hub, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func clientJoinsLobby(w http.ResponseWriter, r *http.Request, hub *Hub, playerID string) error {
	log.Println("UPGRDING")
	log.Println(playerID)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("connection error:", err)
		return errors.New("connection error")
	}

	client := &Client{conn: ws, send: make(chan Message, 256), sendJSON: make(chan JSONMessage, 256)} // Increase buffer size

	if playerID != "" {
		for c := range hub.clients {
			if c.player.ID == playerID {
				client = c
				client.conn = ws
				c.send = make(chan Message)
				c.sendJSON = make(chan JSONMessage)
				log.Println("REASSIGNED CLIENT")
			}
		}
	}

	go client.ReadPump(hub)
	go client.WritePump(hub)

	return nil
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "server/static/favicon.ico")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("static"))
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}
	fs.ServeHTTP(w, r)
}

func gameLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var data struct {
			Name string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		// Use the name value as needed
	}

	tmpl, err := template.ParseFiles("static/gamelobby.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

	connectionHandler(w, r)
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
