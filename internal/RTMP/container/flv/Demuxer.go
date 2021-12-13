package flv

import (
	"fmt"
	"git.hub.com/wangyl/MediaSreamServer/internal/RTMP/container"
)

var (
	ErrAvcEndSEQ = fmt.Errorf("avc end sequence")
)

type DeMuxer struct {
}

func NewDeMuxer() *DeMuxer {
	return &DeMuxer{}
}

func (d *DeMuxer) DeMuxH(p *container.Packet) error {
	var tag Tag
	_, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag
	return nil
}

func (d *DeMuxer) DeMux(p *container.Packet) error {
	var tag Tag
	n, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	if tag.CodecID() == container.VIDEO_H264 &&
		p.Data[0] == 0x17 && p.Data[1] == 0x02 {
		return ErrAvcEndSEQ
	}
	p.Header = &tag
	p.Data = p.Data[n:]

	return nil
}
