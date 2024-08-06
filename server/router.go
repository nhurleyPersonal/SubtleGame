package server

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func connectionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HERE 4", r.URL)
	lobbyID := r.URL.Query().Get("lobbyID")
	playerName := r.URL.Query().Get("name")
	log.Println(lobbyID, playerName)

	// Check if the request is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(r) {
		err := clientJoinsLobby(w, r, lobbyID, playerName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func clientJoinsLobby(w http.ResponseWriter, r *http.Request, lobbyID string, playerName string) error {
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

	hub.register <- client

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
	lobbyID := r.URL.Query().Get("lobbyID")
	playerName := r.URL.Query().Get("name")
	wss := "wss://"
	if r.Host == "localhost:8080" {
		wss = "ws://"
	}
	WsURL := wss + r.Host + "/youmayenter?lobbyID=" + lobbyID + "&name=" + playerName

	tmpl, err := template.ParseFiles("server/templates/gameroom.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	data := struct {
		WsURL      string
		PlayerName string
	}{
		WsURL:      WsURL,
		PlayerName: playerName,
	}
	tmpl.Execute(w, data)
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
