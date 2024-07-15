package testcases

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"multiplayer_game/controller/websocketserver"
	"multiplayer_game/dto"
	"multiplayer_game/service/redispubsub"
	"multiplayer_game/util"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

const (
	numGames   = 10
	joinFactor = 10
)

var (
	gameIdWsClientMap = make(map[string][]*websocket.Conn)
	mu                sync.Mutex
)

// Tests message flow with different testcases in different order
func TestWebsocketMessageFlow(t *testing.T) {
	go redispubsub.SubscribeToRedisChannel(websocketserver.RedisClient, websocketserver.GameIdConnectionIdMap, websocketserver.ConnectionIdConnectionMap)
	s := httptest.NewServer(http.HandlerFunc(websocketserver.HandleConnections))
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	gameIds := []string{}
	var wg sync.WaitGroup
	var wsList []*websocket.Conn
	var laterWsList []*websocket.Conn
	//create numSockets games, check if they are created successfully
	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			ws := createMockWSClient(u, t)
			gameId := testCreateMessage(ws, t)
			mu.Lock()
			wsList = append(wsList, ws)
			gameIds = append(gameIds, gameId)
			gameIdWsClientMap[gameId] = []*websocket.Conn{ws}
		}()
	}

	wg.Wait()

	// create some new sockets, we will use them later to test other message actions
	for i := 0; i < numGames*joinFactor; i++ {
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
	for j := 0; j < numGames*joinFactor; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mu.Unlock()
			randomGame := j % numGames
			randomSocket := j
			testJoinMessage(randomGame, randomSocket, laterWsList, gameIds, t)
			mu.Lock()
			gameIdWsClientMap[gameIds[randomGame]] = append(gameIdWsClientMap[gameIds[randomGame]], laterWsList[randomSocket])
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
	wg.Wait()

	log.Printf("------TESTING----DRAW-------")
	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomGame := i
			testDrawMessage(gameIds[randomGame], t)
		}()
	}
	wg.Wait()
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
	randomText := util.GenerateGuid()
	pattern := fmt.Sprintf("^[a-zA-Z0-9]+ Says %s", randomText)
	var messageFromClient = dto.MessageFromClient{Action: "chat", GameId: game, ChatText: randomText}
	mu.Lock()
	gameWsSockets := gameIdWsClientMap[game]
	mu.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(gameWsSockets))
	for _, v := range gameWsSockets {
		wg.Add(1)
		go func() {
			testListenIncomingChatMessage(v, regexp.MustCompile(pattern), &wg, errChan)
		}()
	}
	randChatter := rand.Intn(len(gameWsSockets))
	gameWsSockets[randChatter].WriteJSON(messageFromClient)

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func testListenIncomingChatMessage(ws *websocket.Conn, regExp *regexp.Regexp, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	for {
		_, resp, err := ws.ReadMessage()
		if err != nil {
			errChan <- fmt.Errorf("error while reading chat message %v", err)
			return
		}

		var messageToClient dto.MessageToClient
		err = json.Unmarshal(resp, &messageToClient)
		if err != nil {
			errChan <- fmt.Errorf("error while parsing chat msg %v %v", string(resp), err)
			return
		}

		if messageToClient.Action != "chat" {
			continue
		}

		if !regExp.MatchString(messageToClient.Data) {
			errChan <- fmt.Errorf("Unexpected chat message %+v", messageToClient)
			return
		}
		break
	}
	errChan <- nil
}

func testDrawMessage(game string, t *testing.T) {
	randomDrawingList := createRandomDrawings()
	mu.Lock()
	gameWsSockets := gameIdWsClientMap[game]
	mu.Unlock()
	var wg sync.WaitGroup
	errChan := make(chan error, len(gameWsSockets))
	for _, v := range gameWsSockets {
		wg.Add(1)
		go func() {
			testListenIncomingDrawMessage(v, randomDrawingList, &wg, errChan)
		}()
	}

	randDrawer := rand.Intn(len(gameWsSockets))

	for _, drawing := range randomDrawingList {
		var messageFromClient = dto.MessageFromClient{Action: "draw", Drawing: drawing, GameId: game}
		gameWsSockets[randDrawer].WriteJSON(messageFromClient)
	}

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func createRandomDrawings() []dto.Drawing {
	var drawings []dto.Drawing

	for i := 0; i < 10; i++ {
		drawings = append(drawings, createRandomDrawing())
	}

	return drawings
}

func createRandomDrawing() dto.Drawing {
	//100 component in draw message
	var drawing dto.Drawing
	for i := 0; i < 100; i++ {
		var drawComponent dto.DrawComponent
		drawComponent.Type = util.GenerateGuid()
		for j := 0; j < 100; j++ {
			point := dto.Point{X: rand.Intn(1000), Y: rand.Intn(1000)}
			drawComponent.Points = append(drawComponent.Points, point)
		}
		drawing.Comoponents = append(drawing.Comoponents, drawComponent)
	}
	return drawing
}

func testListenIncomingDrawMessage(ws *websocket.Conn, drawings []dto.Drawing, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()
	numDrawings := len(drawings)
	for {
		_, resp, err := ws.ReadMessage()
		if err != nil {
			errChan <- fmt.Errorf("error while reading chat message %v", err)
			return
		}

		var messageToClient dto.MessageToClient
		err = json.Unmarshal(resp, &messageToClient)
		if err != nil {
			errChan <- fmt.Errorf("error while parsing chat msg %v %v", string(resp), err)
			return
		}

		if messageToClient.Action != "draw" {
			continue
		}

		numDrawings--
		if numDrawings == 0 {
			break
		}
	}
	errChan <- nil
}
