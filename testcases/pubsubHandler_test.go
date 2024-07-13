package testcases

import (
	"errors"
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

var (
	mockMu                sync.Mutex
	redisClient           *redis.Client
	gameIdConnectionIdMap *map[string][]string
	mRedis                *miniredis.Miniredis
)

func initialise() error {
	// mini redis is a mock library for testing redis
	var err error
	mRedis, err = miniredis.Run()
	if err != nil {
		return errors.New("mini redis instanciation failed")
	}

	redisClient = redis.NewClient(&redis.Options{Addr: mRedis.Addr()})
	return nil
}

// func TestSubscribeToRedis(t *testing.T) {
// 	if initialise() != nil {
// 		t.FailNow()
// 	}
// 	redispubsub.SubscribeToRedisChannel(redisClient, &map[string][]string{}, &map[string]*websocket.Conn{
// 		"conn1": &MockWebSocketConn{},
// 	})
// }
