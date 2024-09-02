package gamedata

import (
	"multiplayer_game/dto"
	"sync"

	"github.com/go-redis/redis"
)

var (
	Mu               sync.Mutex
	GameErrorChan                  = make(chan *dto.GameErrorEvent, 1000)
	SuccessChatChan                = make(chan *dto.SuccessChatEvent, 50000)
	FailChatChan                   = make(chan *dto.FailChatEvent, 100000)
	GameCompleteChan               = make(chan *dto.GameCompleteEvent, 20000)
	GameStartChan                  = make(chan *dto.GameStartEvent, 20000)
	RedisClient      *redis.Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	WordList []string
)
