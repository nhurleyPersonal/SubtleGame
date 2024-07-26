package server

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GameState struct {
	Players map[string]Player `json:"players"`
	Round   int               `json:"round"`
	Started bool              `json:"started"`
	mu      sync.Mutex
}

func NewGameState() GameState {
	return GameState{
		Players: make(map[string]Player),
		Round:   0,
		Started: false,
	}
}

type Player struct {
	Name    string
	Word    string
	Guesses map[string][]string
	ID      string `json:"id"`
	Leader  bool
}

type PassToClientPlayer struct {
	Name   string
	ID     string `json:"id"`
	Leader bool
}

func (gs *GameState) NewPlayer(name string) Player {
	return Player{
		ID:     uuid.New().String(),
		Name:   name,
		Leader: false,
	}
}

func (gs *GameState) StartGame() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Started = true
	gs.Round = 1

	// Start a timer for 15 seconds
	timer := time.NewTimer(15 * time.Second)
	go func() {
		<-timer.C
		gs.mu.Lock()
		defer gs.mu.Unlock()
		gs.Round = 2
	}()
}

func (gs *GameState) SetWord(player *Player, word string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if !gs.Started || gs.Round != 1 {
		return errors.New("Word not set, it needs to be set in the first round.")
	}
	player.Word = word
	return nil
}

func (gs *GameState) JoinGame(name string, c *Client, hub *Hub) (Player, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if len(gs.Players) >= 3 {
		failedToJoinMessage := Message{
			Type: "lobbyFull",
		}

		c.send <- failedToJoinMessage
		return Player{}, errors.New("Lobby is full")
	}
	player := gs.NewPlayer(name)
	gs.Players[player.ID] = player
	c.player = player

	return player, nil

}

func (gs *GameState) GetPlayer(playerId string) *Player {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	player, ok := gs.Players[playerId]
	if !ok {
		return nil
	}
	return &player
}

func (gs *GameState) GetPlayers() []Player {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	players := make([]Player, 0, len(gs.Players))
	for _, player := range gs.Players {
		players = append(players, player)
	}
	return players
}

func (gs *GameState) GetPassToClientPlayers() []PassToClientPlayer {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	players := make([]PassToClientPlayer, 0, len(gs.Players))
	for _, player := range gs.Players {
		players = append(players, PassToClientPlayer{
			Name: player.Name,
			ID:   player.ID,
		})
	}
	return players
}

func (gs *GameState) RemovePlayer(c *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	log.Println("Removing player", c.player)
	_, ok := gs.Players[c.player.ID]
	if !ok {
		return
	}
	delete(gs.Players, c.player.ID)
	log.Printf("Removed player with ID %s", c.player.ID)
}

// func GuessWord(hub *Hub, playerOne *Player, playerTwo *Player, word string) {
// 	hub.mu.Lock()
// 	defer hub.mu.Unlock()
// 	if _, ok := playerOne.Guesses[playerTwo.ID]; !ok {
// 		playerOne.Guesses[playerTwo.ID] = []string{}
// 	}
// 	playerOne.Guesses[playerTwo.ID] = append(playerOne.Guesses[playerTwo.ID], word)

// }
