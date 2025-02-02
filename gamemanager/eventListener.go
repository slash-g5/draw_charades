package gamemanager

import (
	"bufio"
	"fmt"
	"log"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/dao"
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"
	"multiplayer_game/service/redispubsub"
	"os"
)

func FillWordList() {
	wordFile := "/home/shashank/Documents/DrawCharades/config/worldlist.txt"
	file, err := os.Open(wordFile)
	if err != nil {
		log.Fatal("Unable to open word file")
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var words []string
	for fileScanner.Scan() {
		words = append(words, fileScanner.Text())
	}
	gamedata.WordList = words
}

func HandleGameErrorEvent() {
	for {
		gameError := <-gamedata.GameErrorChan
		if gameError.Type == "fatal" {
			//TODO inform redis to broadcast all game's connections to close game
			log.Printf("Error Happend in game = %+v", gameError)
		}
	}
}

func HandleGameChatSuccess() {
	for {
		event := <-gamedata.SuccessChatChan
		log.Printf("Success Chat Event event = %+v", event)
		//Update curr game score to reflect success
		updateCurrRoundScore(event)
		//TODO inform redis pubsub that score of curr round has been updated
		redispubsub.NotifyGeneralMessage(
			event.GameId,
			gamedata.RedisClient,
			websocketserver.GameIdConnectionIdMap,
			websocketserver.ConnectionIdConnectionMap,
			websocketserver.GameIdGameStateMap,
			dto.MessageToClient{Action: "chat",
				Data: event.ChatterId + " guessed the word",
			},
		)
	}
}

func updateCurrRoundScore(event *dto.SuccessChatEvent) {
	//all connections will eventually update ui clients that someone has scored
	defer gamedata.WsConnLock.Unlock()
	fmt.Printf("success chat event received %+v", event)
	gamedata.WsConnLock.Lock()
	var gameState = websocketserver.GameIdGameStateMap[event.GameId]
	for playerId := range gameState.CurrPlayerScoreMap {
		if playerId == event.ChatterId {
			return
		}
	}
	var answeredB4 = len(gameState.CurrPlayerScoreMap)
	var score = (gameState.TotalPlayers/(uint8(answeredB4)+1))*100 + (gameState.DrawTime/(uint8(gameState.CurrTime)+1))*100
	gameState.CurrPlayerScoreMap[event.ChatterId] = uint16(score)
}

func HandleGameChatFail() {
	//TODO shift code here for normal chat
	for {
		event := <-gamedata.FailChatChan
		log.Printf("Fail Chat Event %v", event)
		redispubsub.NotifyGeneralMessage(event.GameId,
			gamedata.RedisClient,
			websocketserver.GameIdConnectionIdMap,
			websocketserver.ConnectionIdConnectionMap,
			websocketserver.GameIdGameStateMap,
			dto.MessageToClient{Action: "chat",
				Chatter: event.ChatterId,
				Data:    event.Text,
			},
		)
	}
}

func HandleGameComplete() {
	for {
		event := <-gamedata.GameCompleteChan
		log.Printf("Game completed %+v", event)
		completeGame(event)
	}
}

func completeGame(event dto.GameCompleteEvent) {
	game, err := dao.GetGameByGameId(event.GameId, gamedata.RedisClient)
	if err != nil {
		log.Printf("Error while completing game %s %+v", event.GameId, err)
		return
	}
	game.Complete = true
	err = dao.UpdateGame(event.GameId, game, gamedata.RedisClient)
	if err != nil {
		log.Printf("Error while completing game in DB %s %+v", event.GameId, err)
	}
	err = redispubsub.NotifyChannel(&dto.MessageToClient{
		Action: "complete",
		GameId: event.GameId,
	}, gamedata.RedisClient)
	if err != nil {
		log.Printf("Error while sending complete message to clients %s %+v",
			event.GameId,
			err)
	}
}
