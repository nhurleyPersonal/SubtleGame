package server

import (
	"errors"
	"log"
	"strings"
	"sync"
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
	Name             string
	Word             string
	Guesses          map[string][]string
	CorrectLetters   map[string]map[string]bool
	PartiallyCorrect map[string]map[string]bool
	GuessScores      map[string]int
	HasFinished      map[string]bool
	Score            int
	ID               string `json:"id"`
	Leader           bool
	Ready            bool
}

type PassToClientPlayer struct {
	Name   string
	ID     string `json:"id"`
	Leader bool
}

func (gs *GameState) NewPlayer(name string, ID string) Player {
	return Player{
		ID:               ID,
		Name:             name,
		Leader:           false,
		Ready:            false,
		Score:            0,
		Guesses:          make(map[string][]string),
		CorrectLetters:   make(map[string]map[string]bool),
		PartiallyCorrect: make(map[string]map[string]bool),
		GuessScores:      make(map[string]int),
		HasFinished:      make(map[string]bool),
	}
}

func (gs *GameState) StartGame() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	for _, player := range gs.Players {
		if !player.Ready {
			return false
		}
	}

	gs.Started = true
	return true
}

func (gs *GameState) SetWord(player *Player, word string) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	ok := SpellChecker.SearchDirect(strings.ToLower(word))
	if !ok {
		return false
	}

	player.Word = word
	player.Ready = true
	gs.Players[player.ID] = *player
	return true
}

func (gs *GameState) JoinGame(name string, playerID string, c *Client, hub *Hub) (Player, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if len(gs.Players) >= 6 {
		failedToJoinMessage := Message{
			Type: "lobbyFull",
		}

		c.send <- failedToJoinMessage
		return Player{}, errors.New("Lobby is full")
	}
	player := gs.NewPlayer(name, playerID)
	gs.Players[player.ID] = player
	c.player = player

	return player, nil

}

func (gs *GameState) ReconnectPlayer(c *Client, hub *Hub) (Player, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	for pID, player := range gs.Players {
		if pID == c.playerID {
			return player, nil
		}
	}

	failedToJoinMessage := Message{
		Type: "reconnectError",
	}

	c.send <- failedToJoinMessage
	return Player{}, errors.New("reconnection failed")

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
		log.Println("PLAYER IN GAMESTATE GETPLAYERS", player.Name)
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
	_, ok := gs.Players[player.ID]
	if !ok {
		return false
	}
	delete(gs.Players, player.ID)
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

	ok := SpellChecker.SearchDirect(strings.ToLower(word))
	if !ok {
		return nil, nil, false
	}

	selfPlayer, ok := gs.Players[selfPlayerID]
	if !ok {
		return nil, nil, false
	}

	targetPlayer, ok := gs.Players[targetPlayerID]
	if !ok {
		return nil, nil, false
	}

	if _, ok := selfPlayer.Guesses[targetPlayer.ID]; !ok {
		selfPlayer.Guesses[targetPlayer.ID] = []string{}
		selfPlayer.GuessScores[targetPlayer.ID] = 1000
		selfPlayer.CorrectLetters[targetPlayer.ID] = make(map[string]bool)
		selfPlayer.PartiallyCorrect[targetPlayer.ID] = make(map[string]bool)
	}

	for _, guess := range selfPlayer.Guesses[targetPlayer.ID] {
		if word == guess {
			return nil, nil, false
		}
	}

	selfPlayer.Guesses[targetPlayer.ID] = append(selfPlayer.Guesses[targetPlayer.ID], word)
	targetWord := targetPlayer.Word
	var rebuiltTarget = targetWord
	var rebuiltGuess = word
	var completelyCorrect []int
	var partiallyCorrect []int

	for i := 0; i < len(word); i++ {
		if word[i] == targetWord[i] {
			completelyCorrect = append(completelyCorrect, i)
			selfPlayer.CorrectLetters[targetPlayer.ID][string(word[i])] = true
			rebuiltTarget = replaceAtIndex(rebuiltTarget, '*', i)
			rebuiltGuess = replaceAtIndex(rebuiltGuess, '*', i)
		}
	}

	for i, guess := range rebuiltGuess {
		for j, target := range rebuiltTarget {
			if guess == target && guess != rune('*') && target != rune('*') {
				partiallyCorrect = append(partiallyCorrect, i)
				selfPlayer.PartiallyCorrect[targetPlayer.ID][string(word[i])] = true
				rebuiltTarget = replaceAtIndex(rebuiltTarget, '*', j)
				rebuiltGuess = replaceAtIndex(rebuiltGuess, '*', i)
			}
		}
	}

	if len(completelyCorrect) == len(targetWord) {
		selfPlayer.Score += selfPlayer.GuessScores[targetPlayer.ID]
		selfPlayer.HasFinished[targetPlayerID] = true
	}

	selfPlayer.GuessScores[targetPlayer.ID] -= (len(partiallyCorrect) * 25) + ((len(targetWord) - len(partiallyCorrect) - len(completelyCorrect)) * 50)

	gs.Players[selfPlayerID] = selfPlayer
	gs.Players[targetPlayerID] = targetPlayer

	return completelyCorrect, partiallyCorrect, true
}

func (gs *GameState) ResetGame() {
	for _, player := range gs.Players {
		player.Score = 0
		player.Word = ""
		player.HasFinished = map[string]bool{}
		player.Guesses = map[string][]string{}
		player.Ready = false
	}
}
