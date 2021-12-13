package cache

import (
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
)

type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (cache *Cache) Write(p *container.Packet) {
	if p.IsMetadata {
		cache.metadata.Write(p)
		return
	} else {
		if !p.IsVideo {
			ah, ok := p.Header.(container.AudioPacketHeader)
			if ok {
				if ah.SoundFormat() == container.SOUND_AAC &&
					ah.AACPacketType() == container.AAC_SEQHDR {
					cache.audioSeq.Write(p)
					return
				} else {
					return
				}
			}

		} else {
			vh, ok := p.Header.(container.VideoPacketHeader)
			if ok {
				if vh.IsSeq() {
					cache.videoSeq.Write(p)
					return
				}
			} else {
				return
			}

		}
	}
	cache.gop.Write(p)
}

func (cache *Cache) Send(w container.HandlePacket) {
	cache.metadata.Send(w)
	cache.videoSeq.Send(w)
	cache.audioSeq.Send(w)

	cache.gop.Send(w)
}
