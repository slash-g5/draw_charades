package gamemanager

import (
	"log"
	"math/rand"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"
	"multiplayer_game/service/redispubsub"
	"sync"
	"time"
)

//Handles repeatative actions for every ongoing game

func HandleRepeatActions() {
	var wg sync.WaitGroup
	for {
		log.Printf("Checking games")
		var gameInSchedule []string
		gamedata.Mu.Lock()
		for gameId := range websocketserver.GameIdConnectionIdMap {
			if _, ok := websocketserver.GameIdGameStateMap[gameId]; !ok {
				log.Printf("State not found in gameId %s", gameId)
				continue
			}
			if !websocketserver.GameIdGameStateMap[gameId].Start {
				//This game is not started
				log.Printf("Found GameId %s but it is not started", gameId)
				continue
			}
			if websocketserver.GameIdGameStateMap[gameId].Complete {
				//Game already completed
				log.Printf("Game already completed %s", gameId)
				continue
			}
			gameInSchedule = append(gameInSchedule, gameId)
		}
		log.Printf("Found gameIds in schedule %v", gameInSchedule)
		gamedata.Mu.Unlock()
		// Unblocking other tasks by releasing lock and sleeping for some time
		time.Sleep(time.Millisecond * 500)

		gamedata.Mu.Lock()
		for _, gameId := range gameInSchedule {
			wg.Add(1)
			go checkCurrGameSchedule(gameId, &wg)
		}
		wg.Wait()
		gamedata.Mu.Unlock()
	}
}

func checkCurrGameSchedule(gameId string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		log.Println(websocketserver.GameIdGameStateMap[gameId])
	}()
	currGameState := websocketserver.GameIdGameStateMap[gameId]

	if currGameState.CurrRound == 0 {
		changeRound(currGameState, gameId)
		redispubsub.NotifyGeneralMessage(gameId,
			gamedata.RedisClient,
			websocketserver.GameIdConnectionIdMap,
			websocketserver.ConnectionIdConnectionMap,
			websocketserver.GameIdGameStateMap,
			dto.MessageToClient{
				Action:        "drawchange",
				Drawer:        currGameState.CurrDrawer,
				Round:         currGameState.CurrRound,
				CurrScoreMap:  currGameState.CurrPlayerScoreMap,
				TotalScoreMap: currGameState.PlayerScoreMap})
		return
	}

	currGameState.CurrTime = uint(time.Since(currGameState.RoundStartTime).Seconds())
	log.Printf("Curr Time = %v \n", currGameState.CurrTime)
	if currGameState.CurrTime < uint(currGameState.RoundTime) {
		return
	}
	if ans, ok := lastDrawerForRound(currGameState, gameId); ok && !ans {
		//Change drawer in same round
		log.Printf("Changing drawer for same round")
		currGameState.AlreadyDrawn = append(currGameState.AlreadyDrawn, currGameState.CurrDrawer)
		currGameState.RoundStartTime = time.Now()
		changeDrawerSame(currGameState, gameId)
		//Notify redispubsub
		redispubsub.NotifyGeneralMessage(gameId,
			gamedata.RedisClient,
			websocketserver.GameIdConnectionIdMap,
			websocketserver.ConnectionIdConnectionMap,
			websocketserver.GameIdGameStateMap,
			dto.MessageToClient{
				Action:        "drawchange",
				Drawer:        currGameState.CurrDrawer,
				Round:         currGameState.CurrRound,
				CurrScoreMap:  currGameState.CurrPlayerScoreMap,
				TotalScoreMap: currGameState.PlayerScoreMap})
		return
	} else if !ok {
		//TODO push to game error chan
		return
	}

	//Copy score from current round to next round
	copyScoreFromCurrRound(currGameState, gameId)

	if currGameState.CurrRound == currGameState.TotalRounds {
		//Complete the game as all rounds are over
		log.Printf("completing game %s as all rounds are over", gameId)
		currGameState.Complete = true
		// Inform redispubsub that game completed
		redispubsub.NotifyGeneralMessage(gameId,
			gamedata.RedisClient,
			websocketserver.GameIdConnectionIdMap,
			websocketserver.ConnectionIdConnectionMap,
			websocketserver.GameIdGameStateMap,
			dto.MessageToClient{
				Action:        "complete",
				TotalScoreMap: currGameState.PlayerScoreMap})
		return
	}
	//Change the round of game
	changeRound(currGameState, gameId)
	//Notify redispubsub
	redispubsub.NotifyGeneralMessage(gameId,
		gamedata.RedisClient,
		websocketserver.GameIdConnectionIdMap,
		websocketserver.ConnectionIdConnectionMap,
		websocketserver.GameIdGameStateMap,
		dto.MessageToClient{
			Action:        "drawchange",
			Drawer:        currGameState.CurrDrawer,
			Round:         currGameState.CurrRound,
			CurrScoreMap:  currGameState.CurrPlayerScoreMap,
			TotalScoreMap: currGameState.PlayerScoreMap})
}

func changeDrawerSame(currGameState *dto.GameState, gameId string) {
	for _, con := range websocketserver.GameIdConnectionIdMap[gameId] {
		if !alreadyDrawn(currGameState, con) {
			currGameState.CurrDrawer = con
			currGameState.CurrWord = genNewWord()
			return
		}
	}
	return
}

func changeRound(currGameState *dto.GameState, gameId string) {
	log.Printf("changing round for game %s", gameId)

	currGameState.CurrRound++
	currGameState.RoundStartTime = time.Now()
	currGameState.CurrTime = 0
	currGameState.AlreadyDrawn = []string{}
	currGameState.CurrDrawer = websocketserver.GameIdConnectionIdMap[gameId][0]
	currGameState.CurrPlayerScoreMap = make(map[string]uint16, 0)
	currGameState.CurrWord = genNewWord()
	return
}

func alreadyDrawn(currGameState *dto.GameState, conId string) bool {
	for _, con := range currGameState.AlreadyDrawn {
		if conId == con {
			return true
		}
	}
	return false
}

func copyScoreFromCurrRound(currGameState *dto.GameState, gameId string) bool {
	log.Println("Copying score to complete current round")
	var connectionIds []string
	var ok bool
	if connectionIds, ok = websocketserver.GameIdConnectionIdMap[gameId]; !ok {
		return false
	}
	for _, conId := range connectionIds {
		if _, ok := currGameState.PlayerScoreMap[conId]; !ok {
			currGameState.PlayerScoreMap[conId] = uint16(0)
		}
		if _, ok := currGameState.CurrPlayerScoreMap[conId]; !ok {
			currGameState.CurrPlayerScoreMap[conId] = uint16(0)
		}
		currGameState.PlayerScoreMap[conId] += currGameState.CurrPlayerScoreMap[conId]
	}
	currGameState.CurrPlayerScoreMap = make(map[string]uint16, 0)
	return true
}

func lastDrawerForRound(currGameState *dto.GameState, gameId string) (bool, bool) {
	if _, ok := websocketserver.GameIdConnectionIdMap[gameId]; !ok {
		return true, false
	}
	var numPlayers = len(websocketserver.GameIdConnectionIdMap[gameId])
	var alreadyDrawn = len(currGameState.AlreadyDrawn)
	if alreadyDrawn >= numPlayers-1 {
		return true, true
	}
	return false, true
}

func genNewWord() string {
	wIndex := rand.Intn(len(gamedata.WordList))
	log.Printf("Generating new word %s", gamedata.WordList[wIndex])
	return gamedata.WordList[wIndex]
}
