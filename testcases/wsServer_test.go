package testcases

import (
	"encoding/json"
	"log"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/dto"
	"multiplayer_game/service/redispubsub"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

const (
	numGames = 100
)

var (
	gameIdWsClientMap map[string]*websocket.Conn
	mu                sync.Mutex
)

func TestWebsocketMessageFlow(t *testing.T) {
	go redispubsub.SubscribeToRedisChannel(websocketserver.RedisClient, &websocketserver.GameIdConnectionIdMap, &websocketserver.ConnectionIdConnectionMap)
	s := httptest.NewServer(http.HandlerFunc(websocketserver.HandleConnections))
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	gameIds := []string{}
	var wg sync.WaitGroup
	var wsList []*websocket.Conn
	var laterWsList []*websocket.Conn
	//create numSockets games, check if they are stored successfully
	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			mu.Lock()
			wsList = append(wsList, createMockWSClient(u, t))
			gameIds = append(gameIds, testCreateMessage(wsList[len(wsList)-1], t))
		}()
	}

	wg.Wait()

	// create some new sockets, we will use them later to test other message actions
	for i := 0; i < numGames*10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			mu.Lock()
			laterWsList = append(laterWsList, createMockWSClient(u, t))
		}()
	}
	wg.Wait()
	log.Printf("------TESTING----JOIN-------")
	// test join action
	//initialise rand to create random numbers
	for j := 0; j < numGames*10; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomGame := j % numGames
			randomSocket := j
			testJoinMessage(randomGame, randomSocket, laterWsList, gameIds, t)
		}()
	}
	wg.Wait()
	log.Printf("------TESTING----CHAT-------")
	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomGame := i
			testChatMessage(gameIds[randomGame], t)
		}()
	}
	//random code to prevent garbage collection
	log.Println(len(wsList))
	log.Println(len(laterWsList))
}

func createMockWSClient(u string, t *testing.T) *websocket.Conn {

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)

	if err != nil {
		t.Fatalf("error while creating websocket")
	}
	return ws
}

func testCreateMessage(ws *websocket.Conn, t *testing.T) string {
	var messageFromClient = dto.MessageFromClient{Action: "create"}
	// matches strings having game with 32 character ids
	pattern := `^CREATED game = ([a-fA-F0-9]{32})$`
	expression := regexp.MustCompile(pattern)
	ws.WriteJSON(messageFromClient)

	_, resp, err := ws.ReadMessage()

	if err != nil {
		t.Fatalf("websocket create message test failed")
	}

	matches := expression.FindStringSubmatch(string(resp))
	if len(matches) < 2 {
		t.Fatalf("websocket create message is incorrectly formatted")
	}

	return matches[1]
}

func testJoinMessage(randomGame int, randomSocket int, wsList []*websocket.Conn, gameIds []string, t *testing.T) {
	var messageFromClient = dto.MessageFromClient{Action: "join", GameId: gameIds[randomGame]}
	wsList[randomSocket].WriteJSON(messageFromClient)

	_, resp, err := wsList[randomSocket].ReadMessage()
	if err != nil {
		t.Fatalf("error while sending join message")
	}

	var messageToClient dto.MessageToClient
	err = json.Unmarshal(resp, &messageToClient)
	if err != nil || messageToClient.Action != "join" {
		t.Fatalf("Incorret message sent to client message = %v resp = %s error = %v", messageToClient, string(resp), err)
	}
	pattern := `^[a-zA-Z0-9]+ joined`
	expression := regexp.MustCompile(pattern)
	if !expression.MatchString(messageToClient.Data) {
		t.Fatalf("Unexpected join response %+v", messageToClient)
	}
}

func testChatMessage(game string, t *testing.T) {

}
