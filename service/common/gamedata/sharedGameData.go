package gamedata

import (
	"multiplayer_game/dto"
	"sync"

	"github.com/go-redis/redis"
)

var (
	RoundChangeChan                   = make(chan dto.GameScheduledEvent, 10000)
	RoundChangeQuitChan               = make(chan bool)
	GameErrorChan                     = make(chan *dto.GameErrorEvent, 1000)
	SuccessChatChan                   = make(chan *dto.SuccessChatEvent, 1000)
	FailChatChan                      = make(chan *dto.FailChatEvent, 1000)
	GameCompleteChan                  = make(chan dto.GameCompleteEvent, 1000)
	GameStartChan                     = make(chan dto.GameStartEvent, 1000)
	QuitGameStartChan                 = make(chan bool)
	GameSchedulerChan                 = make(chan *dto.GameScheduledEvent, 1000)
	scheduledGames                    = make([]string, 0)
	RedisClient         *redis.Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	WordList []string
	// Lock for scheduled Games
	sgLock sync.Mutex
	// Lock for websocket connections
	WsConnLock sync.Mutex
)

func AddGameToSchedule(gameId string) {
	sgLock.Lock()
	defer sgLock.Unlock()
	scheduledGames = append(scheduledGames, gameId)
}

func RemoveGameToSchedule(gameId string) {
	sgLock.Lock()
	defer sgLock.Unlock()
	newList := make([]string, 0)
	for _, v := range scheduledGames {
		if v == gameId {
			continue
		}
		newList = append(newList, v)
	}
	scheduledGames = newList
}

func GetScheduledGames() []string {
	sgLock.Lock()
	defer sgLock.Unlock()
	tempList := make([]string, 0)
	for _, v := range scheduledGames {
		tempList = append(tempList, v)
	}
	return tempList
}
