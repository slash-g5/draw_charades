package main

import (
	"multiplayer_game/controller/httpserver"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/gamemanager"
)

var (
	numWorkers  = 6
	numChannels = 1000
)

func main() {
	go gamemanager.FillWordList()
	go gamemanager.HandleGameComplete()
	go gamemanager.HandleGameChatSuccess()
	go gamemanager.HandleRepeatActions()
	go httpserver.Initialize()
	go func() {
		dispatcher := gamemanager.NewDispatcher(numWorkers)
		dispatcher.Run(numWorkers, numChannels)
	}()
	websocketserver.Initialize()
}
