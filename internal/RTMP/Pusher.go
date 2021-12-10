package RTMP

import (
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"github.com/gwuhaolin/livego/av"
	"sync"
	"time"
)

type PusherManager struct {
	pusherLock *sync.RWMutex
	pusher     map[string]*Pusher
}

func NewPusherManager() *PusherManager {
	return &PusherManager{
		pusherLock: new(sync.RWMutex),
		pusher:     make(map[string]*Pusher),
	}
}

func (pThis *PusherManager) addPusher(pusher *Pusher) (old *Pusher, isExit bool) {
	pThis.pusherLock.Lock()
	defer pThis.pusherLock.Unlock()
	if old, isExit = pThis.pusher[pusher.Id]; isExit {
		return
	} else {
		pThis.pusher[pusher.Id] = pusher
		return pusher, false
	}
}

func (pThis *PusherManager) getPusher(id string) (*Pusher, bool) {
	pThis.pusherLock.RLock()
	old, ok := pThis.pusher[id]
	pThis.pusherLock.RUnlock()
	return old, ok
}

func (pThis *PusherManager) pusherIsExit(id string) (pusher *Pusher, isExit bool) {
	pThis.pusherLock.RLock()
	defer pThis.pusherLock.RUnlock()
	if pusher, isExit = pThis.pusher[id]; isExit {
		return
	} else {
		return nil, false
	}
}

func (pThis *PusherManager) removePusher(pusher *Pusher) {
	pThis.pusherLock.Lock()
	delete(pThis.pusher, pusher.Id)
	pThis.pusherLock.Unlock()
}

type Pusher struct {
	Id string
	s  *Session

	player      map[string]*Player
	playerMutex sync.RWMutex

	stop bool
}

func NewPusher(id string, s *Session) {
	if _, ok := s.srv.PushManager.pusherIsExit(id); ok {
		return
	}
	pusher := &Pusher{
		Id:          id,
		s:           s,
		player:      nil,
		playerMutex: sync.RWMutex{},
	}
	s.srv.PushManager.addPusher(pusher)
	s.StopHandleFunc = append(s.StopHandleFunc, func() {
		pusher.playerMutex.Lock()
		for _, player := range pusher.player {
			player.Stop()
		}
		pusher.playerMutex.Unlock()
	})
	go pusher.SendPacket()
	go pusher.checkPusher()
}

func (pThis *Pusher) SendPacket() {
	defer func() {
		if err := recover(); err != nil {
			Logger.GetLogger().Errorf("Pusher.SendPacket Panic:%v", err)
		}
		pThis.Stop()
	}()
	for !pThis.stop {
		packet, err := pThis.readPacket()
		if err != nil {
			return
		}
		pThis.broadcast(packet)
	}
}

func (pThis *Pusher) checkPusher() {
	defer func() {
		pThis.Stop()
	}()
	for {
		time.Sleep(time.Second * 10)
		pThis.playerMutex.RLock()
		playerNums := len(pThis.player)
		pThis.playerMutex.RUnlock()
		if playerNums == 0 {
			break
		}
	}
}

func (pThis *Pusher) broadcast(packet Packet) {
	pThis.playerMutex.RLock()
	for _, player := range pThis.player {
		player.HandlePacket(packet)
	}
	pThis.playerMutex.RUnlock()
}

func (pThis *Pusher) readPacket() (packet Packet, err error) {
	var cs Chunk
	for {
		cs, err = pThis.s.readMsg()
		if err != nil {
			return
		}
		if cs.typeId == av.TAG_AUDIO ||
			cs.typeId == av.TAG_VIDEO ||
			cs.typeId == av.TAG_SCRIPTDATAAMF0 ||
			cs.typeId == av.TAG_SCRIPTDATAAMF3 {
			break
		}
	}
	packet.IsAudio = cs.typeId == av.TAG_AUDIO
	packet.IsVideo = cs.typeId == av.TAG_VIDEO
	packet.IsMetadata = cs.typeId == av.TAG_SCRIPTDATAAMF0 || cs.typeId == av.TAG_SCRIPTDATAAMF3
	packet.StreamID = cs.streamId
	packet.Data = cs.Data
	packet.TimeStamp = cs.timestamp
	return
}

func (pThis *Pusher) removerPlayer(id string) {
	pThis.playerMutex.Lock()
	delete(pThis.player, id)
	pThis.playerMutex.Unlock()
}

func (pThis *Pusher) addPlayer(player *Player) {
	pThis.playerMutex.Lock()
	pThis.player[player.Id] = player
	pThis.playerMutex.Unlock()
}

func (pThis *Pusher) Stop() {
	if pThis.stop {
		return
	}
	pThis.stop = true
	pThis.s.stop()
}
