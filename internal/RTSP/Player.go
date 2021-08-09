package RTSP

import (
	"encoding/binary"
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTP"
	"os"
	"runtime"
	"sync"
)

type Player struct {
	s          *Session
	cond       *sync.Cond
	queue      []RTP.Frame
	queueLimit int
}

func NewPlayer(pusher *Pusher, s *Session) *Player {
	player := &Player{
		s:    s,
		cond: sync.NewCond(&sync.Mutex{}),
	}
	if old, isExit := pusher.addPlayer(player); isExit {
		return old
	}
	s.stopHandleFunc = append(s.stopHandleFunc, func() {
		player.cond.Broadcast()
		pusher.playerMutex.Lock()
		pusher.removePlayer(s.sessionID)
		pusher.playerMutex.Unlock()
	})
	s.rtpHandleFunc = append(s.rtpHandleFunc, func(frame RTP.Frame) {
		if s.stoped {
			return
		}
		player.cond.L.Lock()
		player.queue = append(player.queue, frame)
		if player.queueLimit > 0 && player.queueLimit < len(player.queue) {
			for i := 0; i < len(player.queue); i++ {
				player.queue = append(player.queue[:i], player.queue[i+1:]...)
			}
		}
		player.cond.Signal()
		player.cond.L.Unlock()
	})
	return player
}

func (pThis *Player) receiverFrame() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1638)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("Panic: %v\n%s\n", err, buf)
			fmt.Fprintf(os.Stderr, pl)
		}
	}()
	for {
		var pack RTP.Frame
		pThis.cond.L.Lock()
		if len(pThis.queue) == 0 {
			pThis.cond.Wait()
		}
		if pThis.s.stoped {
			break
		}
		pack = pThis.queue[0]
		pThis.queue = pThis.queue[1:]
		pThis.cond.L.Unlock()
		var channel int
		switch pack.SendType {
		case RTP_TYPE_VEDIO:
			channel = pThis.s.vChannel
		case RTP_TYPE_VIDEOCONTROL:
			channel = pThis.s.vChannelControl
		case RTP_TYPE_AUDIO:
			channel = pThis.s.aChannel
		case RTP_TYPE_AUDIOCONTROL:
			channel = pThis.s.vChannelControl
		default:
			continue
		}
		//rtpPacket,err :=  RTP.ParseRTPPack(pack.Data)
		//if err != nil {
		//	Logger.GetLogger().Error("Parse Rtp Packet:"+err.Error(), zap.String("ChannelCode", pThis.s.channelCode))
		//	return
		//}
		//fmt.Println(rtpPacket)
		var dataLen = make([]byte, 2)
		binary.BigEndian.PutUint16(dataLen, uint16(pack.DataLen))
		pThis.s.connRwLock.Lock()
		pThis.s.connRW.WriteByte(0x24)
		pThis.s.connRW.WriteByte(byte(channel))
		pThis.s.connRW.Write(dataLen)
		pThis.s.connRW.Write(pack.Data)
		pThis.s.connRW.Flush()
		pThis.s.connRwLock.Unlock()
	}
}

func (pThis *Player) sendFrame(frame RTP.Frame) {
	for _, f := range pThis.s.rtpHandleFunc {
		f(frame)
	}
}

func (pThis *Player) Stop() {
	pThis.s.stop()
}
