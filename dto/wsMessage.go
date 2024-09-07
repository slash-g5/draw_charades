package dto

type MessageFromClient struct {
	Action   string
	GameId   string
	ChatText string
	ClientId string
	Drawing  Drawing
}

type Point struct {
	X, Y int
}

type DrawComponent struct {
	Points []Point
	Type   string
}

type Drawing struct {
	Comoponents []DrawComponent
}

type MessageToClient struct {
	Action        string
	Data          string
	Drawing       Drawing
	Drawer        string
	Chatter       string
	Round         uint8
	CurrScoreMap  map[string]uint16
	TotalScoreMap map[string]uint16
	Word          string
	ConnectionId  string
	GameId        string
}
