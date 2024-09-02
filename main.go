package main

import (
	"multiplayer_game/controller/httpserver"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/gamemanager"
)

func main() {
	go gamemanager.FillWordList()
	go gamemanager.HandleGameStart()
	go gamemanager.HandleGameChatSuccess()
	go gamemanager.HandleRepeatActions()
	go httpserver.Initialize()
	websocketserver.Initialize()
}
