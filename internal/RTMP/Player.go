package RTMP

import (
	"fmt"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"go.uber.org/zap"
)

type Player struct {
	Id string

	s *Session

	init bool
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

func (pThis *Player) HandlePacket(packet *container.Packet) {
	var err error
	defer func() {
		if err != nil {
			pThis.s.StopCodec = fmt.Sprintf("Player HandlePacket Fail:%v", err)
			pThis.Stop()
		}
	}()
	var cs Chunk
	cs.Data = packet.Data
	cs.length = uint32(len(packet.Data))
	cs.streamId = packet.StreamID
	cs.timestamp = packet.TimeStamp

	if packet.IsVideo {
		cs.typeId = container.TAG_VIDEO
	} else {
		if packet.IsMetadata {
			cs.typeId = container.TAG_SCRIPTDATAAMF0
		} else {
			cs.typeId = container.TAG_AUDIO
		}
	}
	err = pThis.s.writeChunk(&cs)
	if err != nil {
		Logger.GetLogger().Error("Write packet fail:"+err.Error(), zap.String("ConnAddr", pThis.s.getAddr()))
		return
	}
}

func (pThis *Player) Check() {
	for {
		if _, err := pThis.s.readMsg(); err != nil {
			Logger.GetLogger().Error("player Check fail:"+err.Error(), zap.String("ConnAddr", pThis.s.getAddr()))
			pThis.s.StopCodec = fmt.Sprintf("Player Check fail:%v", err)
			pThis.Stop()
			return
		}
	}
}

func (pThis *Player) Stop() {
	if pThis.stop {
		return
	}
	pThis.stop = true
	pThis.s.stop()
}
