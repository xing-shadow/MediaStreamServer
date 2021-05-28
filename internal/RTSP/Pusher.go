package RTSP

import (
	"fmt"
	"sync"
)

const QueueLime = 32

type Pusher struct {
	Id    string
	s     *Session
	queue chan []byte

	player      map[string]*Player
	playerMutex sync.RWMutex
}

func NewPusher(s *Session) *Pusher {
	pusher := &Pusher{
		s:      s,
		Id:     s.ChannelCode,
		queue:  make(chan []byte, QueueLime),
		player: make(map[string]*Player),
	}

	s.Server.PushManager.AddPusher(pusher)
	s.RtpHandleFunc = append(s.RtpHandleFunc, func(data []byte) {
		pusher.queue <- data
	})
	s.StopHandleFunc = append(s.StopHandleFunc, func() {
		s.Server.PushManager.Remove(pusher)
		pusher.ClearPlayer()
	})
	go pusher.ReceiveRtp()
	return pusher
}

func (pThis *Pusher) ReceiveRtp() {
	for !pThis.s.Stoped {
		select {
		case rtp, ok := <-pThis.queue:
			if !ok {
				return
			}
			pThis.playerMutex.RLock()
			//TODO send rtp包
			fmt.Println(rtp)
			pThis.playerMutex.RUnlock()
		case <-pThis.s.Exit:
			return
		}
	}
}

func (pThis *Pusher) Stop() {
	pThis.s.Stop()
}

func (pThis *Pusher) ClearPlayer() {
	pThis.playerMutex.Lock()
	defer pThis.playerMutex.Unlock()
	for _, player := range pThis.player {
		player.Stop()
	}
}
