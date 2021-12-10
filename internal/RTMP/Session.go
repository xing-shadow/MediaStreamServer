package RTMP

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"git.hub.com/wangyl/MediaSreamServer/internal/RichConn"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Snowflake"
	"github.com/gwuhaolin/livego/protocol/amf"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	sessionID     string
	transactionID int
	streamID      int
	connInfo      ConnectInfo
	publishInfo   PublishInfo
	isPublisher   bool

	ctx        *Context
	srv        *RtmpServer
	richConn   *RichConn.ConnRich
	connRw     *RichConn.ReaderWriter
	connRwLock sync.Mutex

	decoder *amf.Decoder
	encoder *amf.Encoder
	buff    *bytes.Buffer

	chunks              map[uint32]Chunk
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	ackReceived         uint32
	done                bool

	StopHandleFunc []func()

	Stoped bool
}

func NewSession(ctx *Context, conn net.Conn, srv *RtmpServer) *Session {
	s := &Session{
		ctx:                 ctx,
		sessionID:           fmt.Sprintf("%d", Snowflake.GenerateId()),
		streamID:            1,
		srv:                 srv,
		richConn:            RichConn.NewConnRich(conn, time.Second*time.Duration(srv.opt.Cfg.ReadTimeout), time.Second*time.Duration(srv.opt.Cfg.WriteTimeout)),
		chunks:              make(map[uint32]Chunk),
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		Stoped:              false,

		decoder: &amf.Decoder{},
		encoder: &amf.Encoder{},
		buff:    bytes.NewBuffer(nil),
	}
	s.connRw = RichConn.NewReadrWriter(s.richConn, 4*1024)
	return s
}

func (s *Session) start() {
	defer func() {
		s.stop()
	}()
	if err := s.handleShake(); err != nil {
		Logger.GetLogger().Error("HandleShake fail:"+err.Error(), zap.String("ConnAddr", s.getAddr()))
		return
	}
	for !s.Stoped {
		if c, err := s.readMsg(); err != nil {
			Logger.GetLogger().Error("Read Msg Error:" + err.Error())
		} else {
			switch c.typeId {
			case 20, 17:
				if err := s.handleCmdMsg(&c); err != nil {
					Logger.GetLogger().Error("Handle Msg Error:" + err.Error())
				}
			}
			if s.done {
				break
			}
		}
	}
	//
	if s.isPublisher {
		_, name, _ := s.getInfo()
		NewPusher(name, s)
	} else {
		_, name, _ := s.getInfo()
		pusher, ok := s.srv.PushManager.getPusher(name)
		if !ok {
			s.stop()
			return
		}
		NewPlayer(s, pusher)
	}
}

func (s *Session) handleCmdMsg(c *Chunk) error {
	var amfType int
	if c.typeId == 17 {
		amfType = amf.AMF3
	} else {
		amfType = amf.AMF0
	}
	r := bytes.NewReader(c.Data)
	vs, err := s.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}
	Logger.GetLogger().Debug(fmt.Sprintf("rtmp req: %#v", vs))
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case cmdConnect:
			if err := s.handleConnectCmd(vs[1:]); err != nil {
				return err
			}
			if err := s.connectCmdResp(c); err != nil {
				return err
			}
		case cmdCreateStream:
			if err := s.handleCreateStreamCmd(vs[1:]); err != nil {
				return err
			}
			if err := s.createStreamResp(c); err != nil {
				return err
			}
		case cmdPublish:
			if err = s.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = s.publishResp(c); err != nil {
				return err
			}
			s.done = true
			s.isPublisher = true
			Logger.GetLogger().Debug("handle publish req done")
		case cmdPlay:
			if err = s.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = s.playResp(c); err != nil {
				return err
			}
			s.done = true
			s.isPublisher = false
			Logger.GetLogger().Debug("handle publish req done")
		case cmdFcpublish:
			return s.fcPublish(vs[1:])
		case cmdReleaseStream:
			return s.releaseStream(vs)
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			Logger.GetLogger().Error("no support command= " + vs[0].(string))
		}
	}
	return nil
}

func (s *Session) readMsg() (Chunk, error) {
	var c Chunk
	for {
		b, err := s.connRw.ReadUintBE(1)
		if err != nil {
			return Chunk{}, err
		}
		format := b >> 6
		csId := b & 0x3f
		switch csId {
		case 0:
			b, err = s.connRw.ReadUintBE(1)
			if err != nil {
				return Chunk{}, err
			}
			csId = 64 + b
		case 1:
			if b, err = s.connRw.ReadUintBE(2); err != nil {
				return Chunk{}, err
			}
			csId = 64 + b
		}
		var ok bool
		c, ok = s.chunks[csId]
		if !ok {
			c.csId = csId
			c.format = format
		} else {
			c.format = format
		}
		err = c.readChunk(s.connRw, s.remoteChunkSize)
		if err != nil {
			return Chunk{}, err
		}
		s.chunks[csId] = c
		if c.got {
			break
		}
	}

	s.handleControlMsg(c)

	s.ack(c.length)

	return c, nil
}

func (s *Session) ack(size uint32) {
	s.ackReceived += size
	if s.ackReceived >= s.remoteWindowAckSize {
		chunk := newAckChunk(s.ackReceived)
		s.writeChunk(&chunk)
		s.ackReceived = 0
	}
}

func (s *Session) handleControlMsg(chunk Chunk) {
	if chunk.typeId == uint32(SetChunkSizeMsgType) {
		s.remoteChunkSize = binary.BigEndian.Uint32(chunk.Data)
	} else if chunk.typeId == uint32(WindowAckSizeMsgType) {
		s.remoteWindowAckSize = binary.BigEndian.Uint32(chunk.Data)
	}
}

func (s *Session) handleShake() (err error) {
	var C0C1C2 = make([]byte, 1+1536*2)
	var S0S1S2 = make([]byte, 1+1536*2)
	C0C1 := C0C1C2[:1+1536]
	if _, err = io.ReadFull(s.connRw, C0C1); err != nil {
		return
	}
	if C0C1[0] != 3 {
		err = fmt.Errorf("rtsp handleshake invalid version:%d", C0C1[0])
		return
	}
	zero := binary.BigEndian.Uint32(C0C1[5:9])
	if zero != 0 {
		err = fmt.Errorf("rtsp handleshake invalid zero")
		return
	}
	S0S1S2[0] = 3
	copy(S0S1S2[1536+1:], C0C1[1:])
	if _, err = s.connRw.Write(S0S1S2[:]); err != nil {
		return
	}
	if _, err = io.ReadFull(s.connRw, C0C1C2[1+1536:]); err != nil {
		return
	}
	return
}

func (s *Session) writeMsg(csid, streamId uint32, args ...interface{}) error {
	s.buff.Reset()
	for _, v := range args {
		if _, err := s.encoder.Encode(s.buff, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := s.buff.Bytes()
	c := Chunk{
		format:    0,
		csId:      csid,
		timestamp: 0,
		typeId:    20,
		streamId:  streamId,
		length:    uint32(len(msg)),
		Data:      msg,
	}
	return s.writeChunk(&c)
}

func (s *Session) writeChunk(c *Chunk) (err error) {
	if c.typeId == uint32(SetChunkSizeMsgType) {
		s.chunkSize = binary.BigEndian.Uint32(c.Data)
	}
	if c.typeId == uint32(AudioMsgType) {
		c.csId = 4
	} else if c.typeId == uint32(VideoMsgType) || c.typeId == uint32(DataAMF0MsgType) || c.typeId == uint32(DataAMF3MsgType) {
		c.csId = 6
	}
	var start int
	var first = true
	s.connRwLock.Lock()
	defer s.connRwLock.Unlock()
	for {
		if start >= int(c.length) {
			break
		}
		if first {
			c.format = 0
			first = false
		} else {
			c.format = 3
		}
		err = s.writeChunkHead(c)
		if err != nil {
			return
		}
		inc := int(s.chunkSize)
		if len(c.Data)-start < int(s.chunkSize) {
			inc = len(c.Data) - start
		}
		buf := c.Data[start : start+inc]
		_, err = s.connRw.Write(buf)
		if err != nil {
			return
		}
		start += inc
	}
	return s.connRw.Flush()
}

func (s *Session) writeChunkHead(c *Chunk) (err error) {
	//Chunk Basic Head
	h := c.format << 6
	switch {
	case c.csId < 64:
		h |= c.csId
		err = s.connRw.WriteUintBE(h, 1)
	case c.csId-64 < 256:
		h |= 0
		err = s.connRw.WriteUintBE(h, 1)
		if err != nil {
			break
		}
		err = s.connRw.WriteUintBE(c.csId-64, 1)
	case c.csId-64 < 65536:
		h |= 1
		err = s.connRw.WriteUintBE(h, 1)
		if err != nil {
			break
		}
		err = s.connRw.WriteUintBE(c.csId-64, 2)
	}
	if err != nil {
		return
	}
	//Chunk Msg Head
	ts := c.timestamp
	if c.format == 3 {
		goto TsExtend
	}
	if ts > 0xffffff {
		err = s.connRw.WriteUintBE(0xffffff, 3)
	} else {
		err = s.connRw.WriteUintBE(ts, 3)
	}
	if err != nil {
		return
	}
	if c.format == 2 {
		goto TsExtend
	}
	if c.length > 0xffffff {
		return fmt.Errorf("msg length too big")
	}
	err = s.connRw.WriteUintBE(c.length, 3)
	if err != nil {
		return
	}
	err = s.connRw.WriteUintBE(c.typeId, 1)
	if err != nil {
		return
	}
	if c.format == 1 {
		goto TsExtend
	}
	err = s.connRw.WriteUintLE(c.streamId, 4)
	if err != nil {
		return
	}
TsExtend:
	if ts > 0xffffff {
		err = s.connRw.WriteUintBE(ts, 4)
	}
	return
}

func (s *Session) getInfo() (app string, name string, url string) {
	app = s.connInfo.App
	name = s.publishInfo.Name
	url = s.connInfo.TcUrl + "/" + s.publishInfo.Name
	return
}

func (s *Session) stop() {
	if s.Stoped == true {
		return
	}
	s.Stoped = true
	for _, f := range s.StopHandleFunc {
		f()
	}
	s.richConn.Conn.Close()
}

type PublishInfo struct {
	Type string
	Name string
}

type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

func (s *Session) handleConnectCmd(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return fmt.Errorf("req error")
			}
			s.transactionID = id
		case amf.Object:
			objMap := v.(amf.Object)
			if app, ok := objMap["app"]; ok {
				s.connInfo.App = app.(string)
			}
			if flashVer, ok := objMap["flashVer"]; ok {
				s.connInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := objMap["tcUrl"]; ok {
				s.connInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := objMap["objectEncoding"]; ok {
				s.connInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (s *Session) connectCmdResp(c *Chunk) error {
	chunk := newWindowAckSizeChunk(2500000)
	err := s.writeChunk(&chunk)
	if err != nil {
		return err
	}
	chunk = newSetPeerBandwidth(2500000)
	err = s.writeChunk(&chunk)
	if err != nil {
		return err
	}
	chunk = newSetChunkSize(1024)
	err = s.writeChunk(&chunk)
	if err != nil {
		return err
	}
	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = s.connInfo.ObjectEncoding
	return s.writeMsg(c.csId, c.streamId, "_result", s.transactionID, resp, event)
}

func (s *Session) handleCreateStreamCmd(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			s.transactionID = int(v.(float64))
		case amf.Object:

		}
	}
	return nil
}

func (s *Session) createStreamResp(c *Chunk) error {
	return s.writeMsg(c.csId, c.streamId, "_result", s.transactionID, nil, s.streamID)
}

func (s *Session) publishOrPlay(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				s.publishInfo.Type = v.(string)
			} else if k == 3 {
				s.publishInfo.Name = v.(string)
			}
		case float64:
			s.transactionID = int(v.(float64))
		case amf.Object:

		}
	}
	return nil
}

func (s *Session) publishResp(c *Chunk) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publising."
	return s.writeMsg(c.csId, c.streamId, "onStatus", 0, nil, event)
}

func (s *Session) playResp(c *Chunk) error {
	if err := s.setBegin(); err != nil {
		return err
	}
	if err := s.setRecord(); err != nil {
		return err
	}
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if err := s.writeMsg(c.csId, c.streamId, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if err := s.writeMsg(c.csId, c.streamId, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if err := s.writeMsg(c.csId, c.streamId, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if err := s.writeMsg(c.csId, c.streamId, "onStatus", 0, nil, event); err != nil {
		return err
	}
	return nil
}

func (s *Session) setBegin() error {
	ret := newUserControlMsg(streamBegin, 4)
	for i := 0; i < 4; i++ {
		ret.Data[i+2] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	return s.writeChunk(&ret)
}

func (s *Session) setRecord() error {
	ret := newUserControlMsg(streamIsRecorded, 4)
	for i := 0; i < 4; i++ {
		ret.Data[i+2] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	return s.writeChunk(&ret)
}

func (s *Session) releaseStream(vs []interface{}) error {
	return nil
}

func (s *Session) fcPublish(vs []interface{}) error {
	return nil
}

func (s *Session) getAddr() string {
	return s.richConn.Conn.RemoteAddr().String()
}
