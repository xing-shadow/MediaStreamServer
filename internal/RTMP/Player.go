package RTMP

type Player struct {
	Id string
}

func NewPlayer(id string) *Player {
	player := &Player{Id: id}
	return player
}
