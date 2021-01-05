package internal

import (
	"bufio"
	"net"
	"time"
)

type Session struct {
	SessionID  string
	CustomPath string

	Url        string
	Agent      string
	AuthEnable bool

	Conn          *ConnRich
	ConnRW        *bufio.ReadWriter
	TransportType int

	vChannel        int
	vChannelControl int
	aChannel        int
	aChannelControl int
	vCodec          string
	vControl        string
	aCodec          string
	aControl        string

	Seq      int
	SdpInfos map[string]*SdpInfo

	Stoped bool
}

//func NewSession(conn net.Conn) {
//	session := &Session{}
//}

func (s *Session) NewRtspClientSession(conn net.Conn, agent string) *Session {
	s.Agent = agent
	//Conn
	s.Conn = NewConnRich(conn)
	s.Conn.SetReadTimeout(time.Second * 10)
	s.Conn.SetWriteTimeout(time.Second * 10)
	s.ConnRW = bufio.NewReadWriter(bufio.NewReader(s.Conn), bufio.NewWriter(s.Conn))
	//Channel
	s.vChannel = 0
	s.vChannelControl = 1
	s.aChannel = 2
	s.aChannelControl = 3
	s.TransportType = TRANS_TYPE_TCP
	return s
}

func (s *Session) Stop() {
	if s.Stoped {
		return
	}
	s.Stoped = true
	if s.Conn != nil {
		s.Conn.conn.Close()
		s.Conn = nil
	}
}
