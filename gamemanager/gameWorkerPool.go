package gamemanager

import (
	"log"
	"math/rand"
	"multiplayer_game/dao"
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"
	"multiplayer_game/service/redispubsub"
	"time"
)

// To handle game scheduled events(repeatative events of round change)
type GameRCWorker struct {
	GameRCWorkerPool chan chan dto.GameScheduledEvent
	RoundChangeChan  chan dto.GameScheduledEvent
	Quit             chan bool
}

// To handle game start events, (we are assuming just one thread to be sufficient for this action)
type GameStartWorker struct {
	GameStartChan chan dto.GameStartEvent
	Quit          chan bool
}

func (gsw *GameStartWorker) Start() {
	go func() {
		for {
			select {
			case event := <-gsw.GameStartChan:
				log.Println("Game start event received for " + event.GameId)
				err := dao.StartGame(event.GameId, gamedata.RedisClient)
				if err != nil {
					log.Printf("Error starting game %+v, error %+v", event.GameId, err)
					continue
				}
				gamedata.AddGameToSchedule(event.GameId)
				redispubsub.NotifyChannel(&dto.MessageToClient{Action: "start", GameId: event.GameId}, gamedata.RedisClient)
			case <-gsw.Quit:
				return
			}
		}
	}()
}

func (grcw *GameRCWorker) Start() {
	go func() {
		for {
			grcw.GameRCWorkerPool <- grcw.RoundChangeChan
			select {
			case event := <-grcw.RoundChangeChan:
				log.Println("Game round change event received for " + event.GameId)
				game, err := dao.GetGameByGameId(event.GameId, gamedata.RedisClient)
				if err != nil {
					log.Printf("Error fetching game data for game %s, error %+v", event.GameId, err)
					continue
				}
				if len(game.ActivePlayers) <= 1 {
					log.Printf("Low no. of players in game %s no %+v", event.GameId, game.ActivePlayers)
					continue
				}
				if !drawerChangeNeeded(game) {
					continue
				}
				game.AlreadyDrawn = append(game.AlreadyDrawn, game.CurrDrawer)
				if !roundChangeNeeded(game) {
					changeDrawerSame(game)
					dao.UpdateGame(event.GameId, game, gamedata.RedisClient)
					redispubsub.NotifyChannel(&dto.MessageToClient{Action: "roundsame",
						GameId: event.GameId,
						Drawer: game.CurrDrawer},
						gamedata.RedisClient)
					continue
				} else if game.CurrRound == game.TotalRounds {
					gamedata.RemoveGameToSchedule(event.GameId)
					game.Complete = true
					dao.UpdateGame(event.GameId, game, gamedata.RedisClient)
					redispubsub.NotifyChannel(&dto.MessageToClient{Action: "complete", GameId: event.GameId},
						gamedata.RedisClient)
					gamedata.GameCompleteChan <- dto.GameCompleteEvent(event)
					continue
				} else {
					copyScoreFromCurrRound(game)
					changeRound(game)
					dao.UpdateGame(event.GameId, game, gamedata.RedisClient)
					redispubsub.NotifyChannel(&dto.MessageToClient{Action: "roundchange",
						GameId:        event.GameId,
						Drawer:        game.CurrDrawer,
						TotalScoreMap: game.PlayerScoreMap},
						gamedata.RedisClient)
				}

			case <-grcw.Quit:
				return
			}
		}
	}()
}

func drawerChangeNeeded(game *dto.GameState) bool {
	currTime := int(time.Since(game.DrawStartTime).Seconds())
	return currTime > int(game.DrawTime)-1
}

func roundChangeNeeded(game *dto.GameState) bool {
	for _, con := range game.ActivePlayers {
		if !alreadyDrawn(game, con) {
			return false
		}
	}
	return true
}

func changeDrawerSame(game *dto.GameState) {
	for _, con := range game.ActivePlayers {
		if !alreadyDrawn(game, con) {
			game.CurrDrawer = con
			game.CurrWord = genNewWord()
			game.DrawStartTime = time.Now()
			game.CurrTime = 0
			return
		}
	}
}

func copyScoreFromCurrRound(game *dto.GameState) bool {
	log.Println("Copying score to complete current round")
	var connectionIds = game.ActivePlayers
	for _, conId := range connectionIds {
		if _, ok := game.PlayerScoreMap[conId]; !ok {
			game.PlayerScoreMap[conId] = uint16(0)
		}
		if _, ok := game.CurrPlayerScoreMap[conId]; !ok {
			game.CurrPlayerScoreMap[conId] = uint16(0)
		}
		game.PlayerScoreMap[conId] += game.CurrPlayerScoreMap[conId]
	}
	game.CurrPlayerScoreMap = make(map[string]uint16, 0)
	return true
}

func changeRound(game *dto.GameState) {
	game.CurrRound++
	game.DrawStartTime = time.Now()
	game.CurrTime = 0
	game.AlreadyDrawn = []string{}
	game.CurrDrawer = game.ActivePlayers[0]
	game.CurrPlayerScoreMap = make(map[string]uint16, 0)
	game.CurrWord = genNewWord()
}

func alreadyDrawn(currGameState *dto.GameState, conId string) bool {
	for _, con := range currGameState.AlreadyDrawn {
		if conId == con {
			return true
		}
	}
	return false
}

func genNewWord() string {
	wIndex := rand.Intn(len(gamedata.WordList))
	log.Printf("Generating new word %s", gamedata.WordList[wIndex])
	return gamedata.WordList[wIndex]
}
