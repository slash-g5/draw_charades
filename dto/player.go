package dto

import "encoding/json"

type Player struct {
	Name, AvatarId string
}

func (p *Player) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}
