package server

import (
	"bytes"
	"html/template"
	"log"
)

func SendCardReset(client *Client, targetPlayerID string) bool {
	resetTmpl, err := template.ParseFiles("server/templates/resetPlayerCard.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var resetTpl bytes.Buffer
	resetData := struct {
		ID string
	}{
		ID: targetPlayerID,
	}

	if err := resetTmpl.Execute(&resetTpl, resetData); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	client.sendhtml <- resetTpl.String()

	return true
}

func SendGuessResults(client *Client, targetPlayerID string, lettersMapped []LetterAndClass) bool {
	// Send the guess Results
	tmpl, err := template.ParseFiles("server/templates/guessResults.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var tpl bytes.Buffer
	data := struct {
		TargetPlayerID string
		LetterResults  []LetterAndClass
	}{
		TargetPlayerID: targetPlayerID,
		LetterResults:  lettersMapped,
	}

	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	client.sendhtml <- tpl.String()

	return true
}

// Need name because the entire div needs to be replaced, so need to know target name as well
func SendCorrectGuess(client *Client, targetPlayerID string, targetPlayerName string, guess string) bool {
	correctTmpl, err := template.ParseFiles("server/templates/correctGuessCard.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var correctTpl bytes.Buffer
	correctData := struct {
		Name string
		ID   string
		Word string
	}{
		Name: targetPlayerName,
		ID:   targetPlayerID,
		Word: guess,
	}

	if err := correctTmpl.Execute(&correctTpl, correctData); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	client.sendhtml <- correctTpl.String()

	return true
}
func UpdateStartButton(hub *Hub) bool {
	readyTmpl, err := template.ParseFiles("server/templates/startGameButtonReady.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var readyTpl bytes.Buffer

	if err := readyTmpl.Execute(&readyTpl, nil); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	hub.broadcasthtml <- readyTpl.String()

	return true
}

func SendPlayerJoinLobby(client *Client, playerName string) bool {
	tmpl, err := template.ParseFiles("server/templates/playerJoinsLobby.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var tpl bytes.Buffer
	data := struct {
		PlayerName string
	}{
		PlayerName: playerName,
	}

	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	client.sendhtml <- tpl.String()

	return true
}

func SendLobbyPlayerReady(client *Client, readyPlayerName string) bool {
	tmpl, err := template.ParseFiles("server/templates/playerReadyLobby.html")
	if err != nil {
		log.Println("template parse error:", err)
		return false
	}

	var tpl bytes.Buffer
	data := struct {
		PlayerName string
	}{
		PlayerName: readyPlayerName,
	}

	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Println("template execute error:", err)
		return false
	}

	client.sendhtml <- tpl.String()

	return true
}
