package websocketserver

import (
	"encoding/json"
	"log"
	"multiplayer_game/dto"
	"multiplayer_game/exceptionhandler"
	"multiplayer_game/service/common/gamedata"
	"multiplayer_game/service/redispubsub"
	"multiplayer_game/util"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var (
	//upgrader upgrades http connection to websocket protocol
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	// define couple of maps to store ws information
	ConnectionIdConnectionMap = make(map[string]*websocket.Conn)
	GameIdConnectionIdMap     = make(map[string][]string)
	// redis client used as pubsub
	RedisClient *redis.Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
)

func Init() {

	log.Println("Initialising WS Server")
	go redispubsub.SubscribeToRedisChannel(RedisClient, GameIdConnectionIdMap, ConnectionIdConnectionMap)

	http.HandleFunc("/ws", HandleConnections)
	log.Println("HTTP server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	var gameId string
	if exceptionhandler.HandleErr(err) {
		log.Printf("Error upgrading to websocket %v", err)
		return
	}
	connectionId := util.GenerateGuid()
	gamedata.Mu.Lock()
	ConnectionIdConnectionMap[connectionId] = ws
	gamedata.Mu.Unlock()
	// listen for eternity using for (true) loop
	for {
		var msg []byte
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message Will Close Socket: %v", err)
			gamedata.Mu.Lock()
			delete(ConnectionIdConnectionMap, connectionId)
			delete(GameIdConnectionIdMap, gameId)
			gamedata.Mu.Unlock()
			break
		}
		var payload dto.MessageFromClient
		err = json.Unmarshal(msg, &payload)
		if err != nil {
			log.Printf("Invalid message %v received from client connectionid %s %v", msg, connectionId, err)
			ws.Close()
			return
		}

		if payload.Action == "create" {
			gameId = util.GenerateGuid()
			gamedata.Mu.Lock()
			GameIdConnectionIdMap[gameId] = []string{connectionId}
			ws.WriteMessage(websocket.TextMessage, []byte("CREATED game = "+gameId))
			gamedata.Mu.Unlock()
			payload.GameId = gameId
			continue
		}

		if payload.Action == "join" {
			gameId = payload.GameId
			gamedata.Mu.Lock()
			if !isValidJoin(gameId, connectionId) {
				ws.WriteMessage(websocket.TextMessage, []byte("Unable To Join "+gameId))
				gamedata.Mu.Unlock()
				continue
			}
			GameIdConnectionIdMap[gameId] = append(GameIdConnectionIdMap[gameId], connectionId)
			payload.ClientId = connectionId
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error while marshaling message %+v %+v \n", payload, err)
			}
			redispubsub.PublishToRedisChannel(RedisClient, payloadBytes)
			gamedata.Mu.Unlock()
			continue
		}

		if payload.Action == "chat" {
			gameId = payload.GameId
			gamedata.Mu.Lock()
			if !isValidMessage(gameId, connectionId) {
				log.Printf("Invalid message for chat %v", payload)
				gamedata.Mu.Unlock()
				continue
			}
			payload.ClientId = connectionId
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error while marshalling message %v %v \n", payload, err)
			}
			redispubsub.PublishToRedisChannel(RedisClient, payloadBytes)
			gamedata.Mu.Unlock()
			continue
		}

		if payload.Action == "draw" {
			gameId = payload.GameId
			gamedata.Mu.Lock()
			if !isValidMessage(gameId, connectionId) {
				log.Printf("Invalid message for chat %v", payload)
				gamedata.Mu.Unlock()
				continue
			}
			payload.ClientId = connectionId
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error while marshalling message %v %v \n", payload, err)
			}
			redispubsub.PublishToRedisChannel(RedisClient, payloadBytes)
			gamedata.Mu.Unlock()
			continue
		}
	}

}

func isValidJoin(gameId string, connectionId string) bool {
	connectionIds, ok := GameIdConnectionIdMap[gameId]
	if !ok {
		log.Printf("gameId %v does not exist", gameId)
		return false
	}
	for _, v := range connectionIds {
		if v == connectionId {
			log.Printf("connection %v already exists in game %v", v, gameId)
			return false
		}
	}
	return true
}

func isValidMessage(gameId string, connectionId string) bool {
	connectionIds, ok := GameIdConnectionIdMap[gameId]
	if !ok {
		return false
	}
	for _, v := range connectionIds {
		if v == connectionId {
			return true
		}
	}

	return false
}
