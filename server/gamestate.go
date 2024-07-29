package server

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/makifdb/spellcheck"
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
	Name        string
	Word        string
	Guesses     map[string][]string
	HasFinished map[string]bool
	Score       int
	ID          string `json:"id"`
	Leader      bool
	Ready       bool
}

type PassToClientPlayer struct {
	Name   string
	ID     string `json:"id"`
	Leader bool
}

func (gs *GameState) NewPlayer(name string) Player {
	return Player{
		ID:          uuid.New().String(),
		Name:        name,
		Leader:      false,
		Ready:       false,
		Guesses:     make(map[string][]string), // Initialize Guesses map
		HasFinished: make(map[string]bool),     // Initialize HasFinished map
	}
}

func (gs *GameState) StartGame() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if gs.Started {
		return false
	}
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
	return true
}

func (gs *GameState) SetWord(player *Player, word string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	// Check spelling of a word
	sc, err := spellcheck.New()
	if err != nil {
		fmt.Println(err)
	}

	ok := sc.SearchDirect(strings.ToLower(word))
	if !ok {
		return errors.New("noword")
	}
	player.Word = word
	player.Ready = true
	gs.Players[player.ID] = *player
	return nil
}

func (gs *GameState) JoinGame(name string, c *Client, hub *Hub) (Player, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if len(gs.Players) >= 6 {
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

func (gs *GameState) GetPlayer(playerId string) (*Player, bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	player, ok := gs.Players[playerId]
	if !ok {
		return nil, false
	}
	return &player, true
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

func (gs *GameState) RemovePlayer(player Player) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	log.Println("Removing player", player)
	_, ok := gs.Players[player.ID]
	if !ok {
		return false
	}
	delete(gs.Players, player.ID)
	log.Printf("Removed player with ID %s", player.ID)
	return true
}

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func (gs *GameState) GuessWord(word string, selfPlayerID string, targetPlayerID string) ([]int, []int, bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	selfPlayer, ok := gs.Players[selfPlayerID]
	if !ok {
		return nil, nil, false
	}

	targetPlayer, ok := gs.Players[targetPlayerID]
	if !ok {
		return nil, nil, false
	}

	sc, err := spellcheck.New()
	if err != nil {
		fmt.Println(err)
	}

	ok = sc.SearchDirect(strings.ToLower(word))
	if !ok {
		return nil, nil, false
	}

	if _, ok := selfPlayer.Guesses[targetPlayer.ID]; !ok {
		selfPlayer.Guesses[targetPlayer.ID] = []string{}
	}
	selfPlayer.Guesses[targetPlayer.ID] = append(selfPlayer.Guesses[targetPlayer.ID], word)
	targetWord := targetPlayer.Word
	var rebuiltTarget = targetWord
	var rebuiltGuess = word
	var completelyCorrect []int
	var partiallyCorrect []int

	log.Println("GUESS !!!!", word)

	log.Println("TARGET WORD !!!!", targetWord)
	for i := 0; i < len(word); i++ {
		if word[i] == targetWord[i] {
			log.Println(word, targetWord, i)
			completelyCorrect = append(completelyCorrect, i)
			rebuiltTarget = replaceAtIndex(rebuiltTarget, '*', i)
			rebuiltGuess = replaceAtIndex(rebuiltGuess, '*', i)
		}
	}

	for i, guess := range rebuiltGuess {
		for j, target := range rebuiltTarget {
			if guess == target && guess != rune('*') && target != rune('*') {
				partiallyCorrect = append(partiallyCorrect, i)
				rebuiltTarget = replaceAtIndex(rebuiltTarget, '*', j)
				rebuiltGuess = replaceAtIndex(rebuiltGuess, '*', i)
			}
		}
	}

	if len(completelyCorrect) == len(targetWord) {
		selfPlayer.Score += 1
		selfPlayer.HasFinished[targetPlayerID] = true
	}

	gs.Players[selfPlayerID] = selfPlayer
	gs.Players[targetPlayerID] = targetPlayer

	return completelyCorrect, partiallyCorrect, true
}
