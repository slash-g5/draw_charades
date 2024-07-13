package main

import (
	"log"
	"multiplayer_game/controller/websocketserver"
)

func main() {
	log.Println("HELLO WORLD!")
	websocketserver.Init()
}
