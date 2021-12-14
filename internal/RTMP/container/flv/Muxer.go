package flv

import (
	"fmt"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

var (
	flvHeader = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
)

const (
	headerLen = 11
)

type FlvWriter struct {
	w        *os.File
	fileName string
	buf      []byte

	closed bool
}

func NewFlvWriter(dir string, app, name string) (f *FlvWriter, err error) {
	path := filepath.Join(dir, app, name)
	fileName := fmt.Sprintf("%s_%d.flv", path, time.Now().Unix())
	f = new(FlvWriter)
	f.fileName = fileName
	f.w, err = creatFile(fileName)
	if err != nil {
		return
	}
	f.buf = make([]byte, headerLen)
	f.w.Write(flvHeader)
	bigEndian(f.buf[:4], 0)
	f.w.Write(f.buf[:4])
	return
}

func (pThis *FlvWriter) HandlePacket(packet *container.Packet) {
	if pThis.w == nil {
		Logger.GetLogger().Error("flv writer already closed", zap.String("Path", pThis.fileName))
		return
	}
	typeID := container.TAG_VIDEO
	if !packet.IsVideo {
		if packet.IsMetadata {
			var err error
			typeID = container.TAG_SCRIPTDATAAMF0
			packet.Data, err = amf.MetaDataReform(packet.Data, amf.DEL)
			if err != nil {
				return
			}
		} else {
			typeID = av.TAG_AUDIO
		}
	}
	dataLen := len(packet.Data)
	timestamp := packet.TimeStamp

	preDataLen := dataLen + headerLen
	timestampBase := timestamp & 0xffffff
	timestampExt := uint8(timestamp >> 24 & 0xff)

	pThis.buf[0] = uint8(typeID)
	bigEndian(pThis.buf[1:4], uint64(dataLen))
	bigEndian(pThis.buf[4:7], uint64(timestampBase))
	pThis.buf[7] = timestampExt
	bigEndian(pThis.buf[8:11], uint64(0))
	pThis.w.Write(pThis.buf)
	pThis.w.Write(packet.Data)
	bigEndian(pThis.buf[:4], uint64(preDataLen))
	pThis.w.Write(pThis.buf[:4])
	return
}

func (pThis *FlvWriter) Close() error {
	if pThis.closed {
		return nil
	}
	pThis.closed = true
	if pThis.w != nil {
		return pThis.w.Close()
	}
	return nil
}

func creatFile(fileName string) (f *os.File, err error) {
	dir := filepath.Dir(fileName)
	if _, errDir := os.Stat(dir); os.IsExist(errDir) {
		return os.Create(fileName)
	} else {
		if err = os.MkdirAll(dir, 0755); err == nil {
			return os.Create(fileName)
		} else {
			return
		}
	}
}

func bigEndian(buf []byte, u uint64) {
	if len(buf) > 8 {
		return
	}
	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = uint8(u & 0xff)
		u = u >> 8
	}
}
