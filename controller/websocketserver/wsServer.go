package websocketserver

import (
	"encoding/json"
	"fmt"
	"log"
	"multiplayer_game/dao"
	"multiplayer_game/dto"
	"multiplayer_game/exceptionhandler"
	"multiplayer_game/service/common/gamedata"
	"multiplayer_game/service/redispubsub"
	"multiplayer_game/util"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	//upgrader upgrades http connection to websocket protocol
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	// define three maps to store ws information
	ConnectionIdConnectionMap = make(map[string]*websocket.Conn)
	GameIdConnectionIdMap     = make(map[string][]string)
	GameIdGameStateMap        = make(map[string]*dto.GameState)
)

func Initialize() {

	log.Println("Initialising WS Server")
	go redispubsub.SubscribeToRedisChannel(gamedata.RedisClient, ConnectionIdConnectionMap)

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
	//Send connectionId to clinet
	var messageToClient = &dto.MessageToClient{Action: "connect", ConnectionId: connectionId}
	ws.WriteJSON(messageToClient)
	// listen for eternity using for (true) loop
	for {
		var msg []byte
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message Will Close Socket: %v", err)
			gamedata.Mu.Lock()
			delete(ConnectionIdConnectionMap, connectionId)
			gamedata.Mu.Unlock()

			if gameId != "" {
				err := dao.MakeConnectionIdIncactive(gameId, connectionId, gamedata.RedisClient)
				if err != nil {
					log.Printf("Error while making connectionId inactive %+v", err)
				}
			}

			message := dto.MessageToClient{
				Action: "disconnect",
				GameId: gameId,
			}
			redispubsub.NotifyChannel(&message, gamedata.RedisClient)
			break
		}
		var payload dto.MessageFromClient
		err = json.Unmarshal(msg, &payload)
		if err != nil {
			log.Printf("Invalid message %v received from client connectionid %s %v", msg, connectionId, err)
			ws.Close()
			return
		}
		payload.ClientId = connectionId

		if payload.Action == "create" {
			gameId = util.GenerateGuid()
			player := dto.Player{Name: payload.Name, AvatarId: payload.AvatarId}
			err = ws.WriteJSON(&dto.MessageToClient{
				Action: "create",
				GameId: gameId,
			})
			if err != nil {
				log.Printf("%+v", err)
				return
			}
			err = dao.AddPlayerByConnectionId(&player, connectionId, gamedata.RedisClient)
			if err != nil {
				log.Printf("%+v", err)
				return
			}
			err = dao.AddPlayerByGameId(&player, gameId, connectionId, gamedata.RedisClient, false)
			if err != nil {
				log.Printf("%+v", err)
				return
			}
			continue
		}

		if payload.Action == "join" {
			gameId = payload.GameId
			player := dto.Player{Name: payload.Name, AvatarId: payload.AvatarId}

			err = dao.AddPlayerByConnectionId(&player, connectionId, gamedata.RedisClient)
			if err != nil {
				fmt.Printf("%+v", err)
				return
			}
			err = dao.AddPlayerByGameId(&player, gameId, connectionId, gamedata.RedisClient, true)
			if err != nil {
				fmt.Printf("%+v", err)
				return
			}
			message := dto.MessageToClient{
				Action:       "join",
				ConnectionId: connectionId,
				GameId:       gameId,
				Name:         player.Name,
			}
			err := redispubsub.NotifyChannel(&message, gamedata.RedisClient)
			if err != nil {
				fmt.Printf("%+v", err)
				return
			}
			continue
		}

		if payload.Action == "chat" {
			gameId = payload.GameId
			gameState, err := dao.GetGameByGameId(gameId, gamedata.RedisClient)
			if err != nil {
				log.Printf("Unable to obtain gamestate for gameId %v", gameId)
				continue
			}
			if gameState.CurrDrawer == payload.ClientId {
				continue
			}
			if payload.ChatText == gameState.CurrWord {
				log.Printf("Successfull guess gameId %s Chat %s", gameId, payload.ChatText)
				gamedata.SuccessChatChan <- &dto.SuccessChatEvent{GameId: gameId, ChatterId: payload.ClientId}
				continue
			}

			message := dto.MessageToClient{
				Action:       "chat",
				ConnectionId: connectionId,
				GameId:       gameId,
				Chatter:      connectionId,
				Data:         payload.ChatText,
			}

			redispubsub.NotifyChannel(&message, gamedata.RedisClient)
			continue
		}

		if payload.Action == "draw" {
			gameId = payload.GameId
			gameState, err := dao.GetGameByGameId(gameId, gamedata.RedisClient)
			if err != nil {
				log.Printf("Unable to obtain gamestate for gameId %v", gameId)
				continue
			}
			if gameState.Start && gameState.CurrDrawer != payload.ClientId {
				log.Printf("Not a drawer hence ignoring message")
				continue
			}
			message := dto.MessageToClient{
				Action: "draw",
				Data:   payload.Data,
				GameId: gameId,
			}
			redispubsub.NotifyChannel(&message, gamedata.RedisClient)
			continue
		}

		if payload.Action == "start" {
			gameId = payload.GameId
			gamedata.Mu.Lock()
			gameState, ok := GameIdGameStateMap[gameId]
			if !ok {
				log.Printf("Invalid message received %+v, invalid game id %+v", payload, gameId)
				gamedata.Mu.Unlock()
				continue
			}
			if gameState.CurrRound > 0 {
				log.Printf("Invalid message received %+v, game already started %s", payload, gameId)
				gamedata.Mu.Unlock()
				continue
			}
			var event = &dto.GameStartEvent{GameId: gameId}
			gamedata.GameStartChan <- event
			gamedata.Mu.Unlock()
			continue
		}
	}

}
