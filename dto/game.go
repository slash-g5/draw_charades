package dto

import (
	"encoding/json"
	"time"
)

type GameState struct {
	TotalRounds        uint8
	CurrRound          uint8
	CurrDrawer         string
	Complete           bool
	AlreadyDrawn       []string
	GameStartTime      time.Time
	Start              bool
	RoundStartTime     time.Time
	RoundTime          uint8
	CurrTime           uint
	ActivePlayers      []string
	InactivePlayers    []string
	TotalPlayers       uint8
	PlayerScoreMap     map[string]uint16
	CurrPlayerScoreMap map[string]uint16
	CurrWord           string
}

// indicates successful chat message
type SuccessChatEvent struct {
	GameId    string
	ChatterId string
}

// indicates wrong guess chat message
type FailChatEvent struct {
	GameId    string
	ChatterId string
	Text      string
}

type GameCompleteEvent struct {
	GameId string
}

type GameErrorEvent struct {
	Type   string
	GameId string
}

type GameStartEvent struct {
	GameId string
}

func (g *GameState) MarshalBinary() ([]byte, error) {
	return json.Marshal(g)
}
