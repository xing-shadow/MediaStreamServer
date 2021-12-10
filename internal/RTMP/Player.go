package RTMP

import (
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"github.com/gwuhaolin/livego/av"
	"go.uber.org/zap"
)

type Player struct {
	Id string

	s    *Session
	stop bool
}

func NewPlayer(s *Session, pusher *Pusher) *Player {
	player := &Player{
		Id: s.sessionID,
		s:  s,
	}
	pusher.addPlayer(player)
	s.StopHandleFunc = append(s.StopHandleFunc, func() {
		pusher.removerPlayer(player.Id)
	})
	return player
}

func (pThis *Player) HandlePacket(packet Packet) {
	var err error
	defer func() {
		if err != nil {
			pThis.Stop()
		}
	}()
	var cs Chunk
	cs.Data = packet.Data
	cs.length = uint32(len(packet.Data))
	cs.streamId = packet.StreamID
	cs.timestamp = packet.TimeStamp

	if packet.IsVideo {
		cs.typeId = av.TAG_VIDEO
	} else {
		if packet.IsMetadata {
			cs.typeId = av.TAG_SCRIPTDATAAMF0
		} else {
			cs.typeId = av.TAG_AUDIO
		}
	}
	err = pThis.s.writeChunk(&cs)
	if err != nil {
		Logger.GetLogger().Error("Write packet fail:"+err.Error(), zap.String("ConnAddr", pThis.s.getAddr()))
		return
	}
}

func (pThis *Player) Stop() {
	if pThis.stop {
		return
	}
	pThis.stop = true
	pThis.s.stop()
}
