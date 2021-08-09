package RTMP

import (
	"encoding/binary"
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RichConn"
)

type Chunk struct {
	format    uint32
	csId      uint32
	timestamp uint32
	length    uint32
	typeId    uint32
	streamId  uint32
	timeDelta uint32
	extend    bool //时间戳是否扩展

	//control
	remain uint32
	index  uint32
	got    bool
	Data   []byte
}

func (c *Chunk) reInit() {
	c.remain = 0
	c.index = 0
	c.got = false
	c.Data = make([]byte, c.length)
}

func (c *Chunk) readChunk(r *RichConn.ReaderWriter, chunkSize uint32) (err error) {
	if c.format != 3 && c.remain != 0 {
		err = fmt.Errorf("invalid remain = %d", c.remain)
		return
	}
	switch c.format {
	case 0: //11 bytes 消息第一块
		c.timestamp, err = r.ReadUintBE(3)
		if err != nil {
			return err
		}
		c.length, err = r.ReadUintBE(3)
		if err != nil {
			return
		}
		c.typeId, err = r.ReadUintBE(1)
		if err != nil {
			return err
		}
		c.streamId, err = r.ReadUintLE(4)
		if err != nil {
			return
		}
		if c.timestamp == 0xfffffff {
			c.timestamp, err = r.ReadUintBE(4)
			if err != nil {
				return
			}
			c.extend = true
		} else {
			c.extend = false
		}
		c.reInit()
	case 1: //7 bytes 消息沿用上一个消息流id,消息长度可能发生改变,
		var timestamp uint32
		timestamp, err = r.ReadUintBE(3) //增量
		if err != nil {
			return err
		}
		c.length, err = r.ReadUintBE(3)
		if err != nil {
			return
		}
		c.typeId, err = r.ReadUintBE(1)
		if err != nil {
			return err
		}
		if timestamp == 0xfffffff {
			timestamp, err = r.ReadUintBE(4)
			if err != nil {
				return
			}
			c.extend = true
		} else {
			c.extend = false
		}
		c.timeDelta = timestamp
		c.timestamp += timestamp
		c.reInit()
	case 2: //3 bytes 时间戳改变
		var timestamp uint32
		timestamp, err = r.ReadUintBE(3) //增量
		if err != nil {
			return err
		}
		if timestamp == 0xfffffff {
			timestamp, err = r.ReadUintBE(4)
			if err != nil {
				return
			}
			c.extend = true
		} else {
			c.extend = false
		}
		c.timeDelta = timestamp
		c.timestamp += timestamp
		c.reInit()
	case 3: //0 bytes 沿用上一个块信息
		if c.remain == 0 {
			switch c.format {
			case 0:
				if c.extend {
					timestamp, _ := r.ReadUintBE(4)
					c.timestamp = timestamp
				}
			case 1, 2:
				var timeDet uint32
				if c.extend {
					timeDet, _ = r.ReadUintBE(4)
				} else {
					timeDet = c.timeDelta
				}
				c.timestamp += timeDet
			}
			c.reInit()
		} else {
			if c.extend { //判断是否发了扩展时间戳
				b, err := r.Peek(4)
				if err != nil {
					return err
				}
				tmpts := binary.BigEndian.Uint32(b)
				if tmpts == c.timestamp {
					r.Discard(4)
				}
			}
		}
	default:
		err = fmt.Errorf("invalid format=%d", c.format)
		return
	}
	size := int(c.remain) // 剩余块大小
	if size > int(chunkSize) {
		size = int(chunkSize)
	}
	if _, err = r.Read(c.Data[c.index : c.index+uint32(size)]); err != nil {
		return
	}
	c.remain -= uint32(size)
	c.index += uint32(size)
	if c.remain == 0 {
		c.got = true
	}
	return
}

func newSetChunkSize(size uint32) Chunk {
	ret := Chunk{
		format:   0,
		csId:     2,
		length:   4,
		typeId:   uint32(SetChunkSizeMsgType),
		streamId: 0,
		Data:     make([]byte, 4),
	}
	binary.BigEndian.PutUint32(ret.Data, size)
	return ret
}

func newAckChunk(size uint32) Chunk {
	ret := Chunk{
		format:   0,
		csId:     2,
		length:   4,
		typeId:   uint32(AckMsgType),
		streamId: 0,
		Data:     make([]byte, 4),
	}
	binary.BigEndian.PutUint32(ret.Data, size)
	return ret
}

func newWindowAckSizeChunk(size uint32) Chunk {
	ret := Chunk{
		format:   0,
		csId:     2,
		typeId:   uint32(WindowAckSizeMsgType),
		streamId: 0,
		length:   4,
		Data:     make([]byte, 4),
	}
	binary.BigEndian.PutUint32(ret.Data, size)
	return ret
}

func newSetPeerBandwidth(size uint32) Chunk {
	ret := Chunk{
		format:   0,
		csId:     2,
		typeId:   uint32(SetPeerBandwidthMsgType),
		streamId: 0,
		length:   5,
		Data:     make([]byte, 5),
	}
	binary.BigEndian.PutUint32(ret.Data[:4], size)
	ret.Data[4] = 2
	return ret
}

func newUserControlMsg(eventType, bufLen uint32) Chunk {
	bufLen += 2
	var ret = Chunk{
		format:   0,
		csId:     2,
		length:   bufLen,
		typeId:   4,
		streamId: 1,
		Data:     make([]byte, bufLen),
	}
	ret.Data[0] = byte(eventType >> 8 & 0xff)
	ret.Data[1] = byte(eventType & 0xff)
	return ret
}
