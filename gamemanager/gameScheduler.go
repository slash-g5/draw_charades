package gamemanager

import (
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"
	"time"
)

//Handles repeatative actions for every ongoing game

func HandleRepeatActions() {
	for {
		for _, gameId := range gamedata.GetScheduledGames() {
			gamedata.RoundChangeChan <- dto.GameScheduledEvent{GameId: gameId}
		}
		time.Sleep(time.Millisecond * 2000)
	}
}
