package RTMP

import (
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/cache"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container/flv"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"go.uber.org/zap"
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
	app string
	Id  string
	s   *Session

	player      map[string]*Player
	playerMutex sync.RWMutex
	flvMuxer    *flv.FlvWriter
	deMuxer     *flv.DeMuxer
	cache       *cache.Cache
	stop        bool
}

func NewPusher(app string, id string, s *Session) (*Pusher, bool) {
	if _, ok := s.srv.PushManager.pusherIsExit(id); ok {
		return nil, false
	}
	pusher := &Pusher{
		app:         app,
		Id:          id,
		s:           s,
		player:      make(map[string]*Player),
		playerMutex: sync.RWMutex{},
		deMuxer:     flv.NewDeMuxer(),
		cache:       cache.NewCache(),
	}
	var err error

	pusher.flvMuxer, err = flv.NewFlvWriter(s.srv.opt.Cfg.FlvDir, app, id)
	if err != nil {
		Logger.GetLogger().Error("create flv file fail:"+err.Error(), zap.String("PusherName", id))
	}
	s.srv.PushManager.addPusher(pusher)
	s.StopHandleFunc = append(s.StopHandleFunc, func() {
		//
		s.srv.PushManager.removePusher(pusher)
		//
		pusher.playerMutex.Lock()
		for _, player := range pusher.player {
			player.s.StopCodec = "player exit ,because pusher exit"
			player.Stop()
		}
		pusher.playerMutex.Unlock()
		//
		pusher.flvMuxer.Close()

	})
	return pusher, true
}

func (pThis *Pusher) SendPacket() {
	for !pThis.stop {
		packet, err := pThis.readPacket()
		if err != nil {
			return
		}
		pThis.cache.Write(&packet)
		if pThis.app == "record" {
			pThis.flvMuxer.HandlePacket(&packet)
		}
		pThis.broadcast(&packet)
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
			pThis.s.StopCodec = "pusher idle time to long,push stop"
			break
		}
	}
}

func (pThis *Pusher) broadcast(packet *container.Packet) {
	pThis.playerMutex.RLock()
	for _, player := range pThis.player {
		if !player.init {
			pThis.cache.Send(player)
			player.init = true
		} else {
			player.HandlePacket(packet)
		}
	}
	pThis.playerMutex.RUnlock()
}

func (pThis *Pusher) readPacket() (packet container.Packet, err error) {
	var cs Chunk
	for {
		cs, err = pThis.s.readMsg()
		if err != nil {
			return
		}
		if cs.typeId == container.TAG_AUDIO ||
			cs.typeId == container.TAG_VIDEO ||
			cs.typeId == container.TAG_SCRIPTDATAAMF0 ||
			cs.typeId == container.TAG_SCRIPTDATAAMF3 {
			break
		}
	}
	packet.IsAudio = cs.typeId == container.TAG_AUDIO
	packet.IsVideo = cs.typeId == container.TAG_VIDEO
	packet.IsMetadata = cs.typeId == container.TAG_SCRIPTDATAAMF0 || cs.typeId == container.TAG_SCRIPTDATAAMF3
	packet.StreamID = cs.streamId
	packet.Data = cs.Data
	packet.TimeStamp = cs.timestamp
	err = pThis.deMuxer.DeMuxH(&packet) //解析flv-tag
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
