package RTP

import (
	"encoding/binary"
	"fmt"
)

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |V=2|P|X|  CC   |M|     PT      |       sequence number         |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                           timestamp                           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           synchronization source (SSRC) identifier            |
   +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
   |            contributing source (CSRC) identifiers             |
   |                             ....                              |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type RTPPack struct {
	Mark        int
	PayloadType int
	Seq         int
	Ts          int
	SSRC        int
	PadLen      int
	Data        []byte
}

func (p RTPPack) String() string {
	return fmt.Sprintf("seq:%v Mark:%v PayLoadType:%v Ts:%v SSRC:%v PadLen:%v", p.Seq, p.Mark, p.PayloadType, p.Ts, p.SSRC, p.PadLen)
}

func ParseRTPPack(src []byte) (RTPPack, error) {
	if len(src) < 12 {
		return RTPPack{}, nil
	}
	p := src[0] >> 5 & 0x1
	cc := src[0] & 0xf
	m := src[1] >> 7 & 0x1
	payload := src[1] & 0x7f
	seq := binary.BigEndian.Uint16(src[2:4])
	ts := binary.BigEndian.Uint32(src[4:8])
	ssrc := binary.BigEndian.Uint32(src[8:12])
	var start, length int
	start = 12 + 4*int(cc)
	var padLen int
	if p != 0 {
		padLen = int(src[len(src)-1] & 0xff)
		if padLen < 0 {
			length = len(src) - 4*int(p) - start
		} else {
			length = len(src) - padLen - start
		}
	} else {
		length = len(src) - 4*int(p) - start
	}
	return RTPPack{
		Mark:        int(m),
		PayloadType: int(payload),
		Seq:         int(seq),
		Ts:          int(ts),
		SSRC:        int(ssrc),
		PadLen:      padLen,
		Data:        src[start : start+length],
	}, nil
}
