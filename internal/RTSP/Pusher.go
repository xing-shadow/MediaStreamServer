package RTSP

import (
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"go.uber.org/zap"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const QueueLime = 32

type Pusher struct {
	Id    string
	s     *Session
	queue chan RTP.Frame

	player      map[string]*Player
	playerMutex sync.RWMutex
	exit        chan struct{}
}

func NewPusher(s *Session) (*Pusher, bool) {
	pusher := &Pusher{
		s:      s,
		Id:     s.channelCode,
		queue:  make(chan RTP.Frame, QueueLime),
		player: make(map[string]*Player),
		exit:   make(chan struct{}),
	}
	if old, isExit := s.Server.PushManager.addPusher(pusher); isExit {
		return old, true
	}
	s.RtpHandleFunc = append(s.RtpHandleFunc, func(frame RTP.Frame) {
		pusher.queue <- frame
	})
	s.StopHandleFunc = append(s.StopHandleFunc, func() {
		close(pusher.exit)
		s.Server.PushManager.remove(pusher)
		pusher.ClearPlayer()
	})
	go pusher.checkPusher()
	return pusher, false
}

func (pThis *Pusher) checkPusher() {
	for !pThis.s.Stoped {
		pThis.playerMutex.RLock()
		var players strings.Builder
		for _, player := range pThis.player {
			players.WriteString(fmt.Sprintf(" %v", player.s.Conn.conn.RemoteAddr().String()))
		}
		pThis.playerMutex.RUnlock()
		Logger.GetLogger().Info("Current players:"+players.String(), zap.String("channelCode", pThis.s.channelCode))
		time.Sleep(time.Second * 15)
	}

}

func (pThis *Pusher) addPlayer(player *Player) (old *Player, isExit bool) {
	pThis.playerMutex.Lock()
	defer pThis.playerMutex.Unlock()
	if old, isExit = pThis.player[player.s.sessionID]; isExit {
		return
	} else {
		pThis.player[player.s.sessionID] = player
		return player, false
	}
}

func (pThis *Pusher) getPlayer(id string) (player *Player, isExit bool) {
	pThis.playerMutex.RLock()
	defer pThis.playerMutex.RUnlock()
	if player, isExit = pThis.player[id]; isExit {
		return
	} else {
		return nil, false
	}
}

func (pThis *Pusher) removePlayer(id string) {
	pThis.playerMutex.Lock()
	defer pThis.playerMutex.Unlock()
	delete(pThis.player, id)
}

func (pThis *Pusher) getSdp() string {
	return pThis.s.sdpRaw
}

func (pThis *Pusher) ReceiveRtp() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1638)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("Panic: %v\n%s\n", err, buf)
			fmt.Fprintf(os.Stderr, pl)
		}
	}()
	for !pThis.s.Stoped {
		select {
		case frame, ok := <-pThis.queue:
			if !ok {
				return
			}
			pThis.playerMutex.RLock()
			for _, player := range pThis.player {
				player.sendFrame(frame)
			}
			pThis.playerMutex.RUnlock()
		case <-pThis.exit:
			return
		}
	}
}

func (pThis *Pusher) ClearPlayer() {
	pThis.playerMutex.Lock()
	defer pThis.playerMutex.Unlock()
	for _, player := range pThis.player {
		player.Stop()
	}
}