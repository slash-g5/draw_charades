package dto

type MessageFromClient struct {
	Action    string
	GameId    string
	ChatText  string
	JoinerId  string
	ChatterId string
}

type MessageToClient struct {
	Action string
	Data   string
}
