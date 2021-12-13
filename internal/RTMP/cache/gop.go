package cache

import "git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"

type GopCache struct {
}

func NewGopCache() *GopCache {
	return &GopCache{}
}

func (pThis *GopCache) Write(packet *container.Packet) {

}

func (pThis *GopCache) Send(packet container.HandlePacket) {

}
