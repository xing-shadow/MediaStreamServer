package RTMP

type Pusher struct {
	Id string
}

func NewPusher(id string) *Pusher {
	pusher := &Pusher{Id: id}
	return pusher
}
