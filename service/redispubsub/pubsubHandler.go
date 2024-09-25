package redispubsub

import (
	"encoding/json"
	"fmt"
	"log"
	"multiplayer_game/dao"
	"multiplayer_game/dto"
	"multiplayer_game/service/common/gamedata"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var (
	redisChannel = "rpubsub"
)

func SubscribeToRedisChannel(redisClient *redis.Client, connectionIdConnectionMap map[string]*websocket.Conn) {
	// subscribe to channel = redisChannel
	pubsub := redisClient.Subscribe(redisChannel)

	ch := pubsub.Channel()
	var message dto.MessageToClient

	for msg := range ch {
		err := json.Unmarshal([]byte(msg.Payload), &message)
		if err != nil {
			log.Printf("Error converting msg payload in format \n")
			continue
		}
		// get the connectionIds associated with the gameId
		game, err := dao.GetGameByGameId(message.GameId, redisClient)
		if err != nil {
			log.Printf("Error while getting game from gameId %s %+v", message.GameId, err)
			continue
		}
		connectionIds := game.ActivePlayers
		// for each connection send message to websocket
		for _, connectionId := range connectionIds {
			sendMessageToWS(connectionId, connectionIdConnectionMap, message)
		}
	}
}

func PublishToRedisChannel(redisClient *redis.Client, msg []byte) {
	err := redisClient.Publish(redisChannel, msg).Err()
	if err != nil {
		log.Printf("Error publishing message to Redis: msg = %v err = %v", msg, err)
	}
}

func sendMessageToWS(connectionId string, connectionIdConnectionMap map[string]*websocket.Conn, message dto.MessageToClient) {
	defer gamedata.Mu.Unlock()
	gamedata.Mu.Lock()
	conn, ok := connectionIdConnectionMap[connectionId]
	if !ok {
		log.Printf("Invalid ConnectionId %v", connectionId)
		return
	}
	err := conn.WriteJSON(message)
	if err != nil {
		log.Printf("Error While writing to websocket %+v", err)
		return
	}
}

func NotifyGameStart(gameId string, redisClient *redis.Client, gameIdConnectionIdMap map[string][]string, connectionIdConnectionMap map[string]*websocket.Conn) {
	log.Printf("Notifying game %s start", gameId)
	defer gamedata.Mu.Unlock()
	gamedata.Mu.Lock()
	connectionIds, ok := gameIdConnectionIdMap[gameId]
	if !ok {
		fmt.Printf("Got no connection id for game %s", gameId)
		return
	}
	for _, con := range connectionIds {
		ws, ok := connectionIdConnectionMap[con]
		if !ok {
			fmt.Printf("Got no connection for connection id %s gameId %s", con, gameId)
			continue
		}
		ws.WriteJSON(dto.MessageToClient{Action: "start", Data: "Game starting now"})
	}
}

func NotifyGeneralMessage(gameId string, redisClient *redis.Client, gameIdConnectionIdMap map[string][]string, connectionIdConnectionMap map[string]*websocket.Conn, gameIdGameStateMap map[string]*dto.GameState, messageTo dto.MessageToClient) {
	log.Printf("Notifying game %s round change", gameId)

	gameState, ok := gameIdGameStateMap[gameId]
	if !ok {
		log.Printf("Invalid Game State for gameId %s", gameId)
	}

	connectionIds, ok := gameIdConnectionIdMap[gameId]
	if !ok {
		log.Printf("Got no connection for game %s", gameId)
		return
	}
	for _, con := range connectionIds {
		ws, ok := connectionIdConnectionMap[con]
		if !ok {
			log.Printf("Got no connection for connection id %s gameId %s", con, gameId)
			continue
		}
		if messageTo.Drawer == con {
			messageTo.Word = gameState.CurrWord
		} else {
			messageTo.Word = ""
		}
		ws.WriteJSON(messageTo)
	}
}

func NotifyChannel(messageTo *dto.MessageToClient, redisClient *redis.Client) error {
	return redisClient.Publish(redisChannel, messageTo).Err()
}
