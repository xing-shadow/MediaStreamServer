package cache

import (
	"bytes"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
	"github.com/gwuhaolin/livego/protocol/amf"
	"log"
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

type SpecialCache struct {
	full bool
	p    *container.Packet
}

func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

func (specialCache *SpecialCache) Write(p *container.Packet) {
	specialCache.p = p
	specialCache.full = true
}

func (specialCache *SpecialCache) Send(w container.HandlePacket) {
	if !specialCache.full {
		return
	}
	w.HandlePacket(specialCache.p)
}
