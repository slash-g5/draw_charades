package redispubsub

import (
	"encoding/json"
	"errors"
	"log"
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var (
	redisChannel = "rpubsub"
)

func SubscribeToRedisChannel(redisClient *redis.Client, gameIdConnectionIdMap map[string][]string, connectionIdConnectionMap map[string]*websocket.Conn) {
	// subscribe to channel = redisChannel
	pubsub := redisClient.Subscribe(redisChannel)

	ch := pubsub.Channel()
	var message dto.MessageFromClient

	for msg := range ch {
		err := json.Unmarshal([]byte(msg.Payload), &message)
		if err != nil {
			log.Printf("Error converting msg payload in format \n")
		}
		gamedata.Mu.Lock()
		connectionIds, ok := gameIdConnectionIdMap[message.GameId]
		gamedata.Mu.Unlock()
		if !ok {
			log.Printf("Invalid gameid received in pubsub %s", message.GameId)
		}
		gamedata.Mu.Lock()
		for _, connectionId := range connectionIds {
			handleMessageWithWSConnection(connectionId, connectionIdConnectionMap, message)
		}
		gamedata.Mu.Unlock()
	}
}

func PublishToRedisChannel(redisClient *redis.Client, msg []byte) {
	err := redisClient.Publish(redisChannel, msg).Err()
	if err != nil {
		log.Printf("Error publishing message to Redis: msg = %v err = %v", msg, err)
	}
}
func handleMessageWithWSConnection(connectionId string, connectionIdConnectionMap map[string]*websocket.Conn, message dto.MessageFromClient) {
	conn, ok := connectionIdConnectionMap[connectionId]
	if !ok {
		log.Printf("Invalid ConnectionId %v", connectionId)
		return
	}

	var err error
	var messageTo dto.MessageToClient = dto.MessageToClient{Action: message.Action}

	if message.Action == "join" {
		messageTo.Data = message.ClientId + " joined"
		err = conn.WriteJSON(messageTo)
	} else if message.Action == "chat" {
		messageTo.Data = message.ClientId + " Says " + message.ChatText
		err = conn.WriteJSON(messageTo)
	} else if message.Action == "draw" {
		messageTo.Drawing = message.Drawing
		err = conn.WriteJSON(messageTo)
	} else {
		err = errors.New("invalid action")
	}

	if err != nil {
		log.Printf("Error While writing to websocket %+v", err)
		return
	}
}
