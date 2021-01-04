package internal

import (
	"bufio"
	"net"
	"time"
)

type Session struct {
	SessionID  string
	CustomPath string

	Url      string
	Agent    string
	authLine string

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

func NewSession(conn net.Conn) {
	session := &Session{}
}

func NewRtspClientSession(conn net.Conn, agent string) *Session {
	session := new(Session)
	session.Agent = agent
	//Conn
	session.Conn = NewConnRich(conn)
	session.Conn.SetReadTimeout(time.Second * 10)
	session.Conn.SetWriteTimeout(time.Second * 10)
	session.ConnRW = bufio.NewReadWriter(bufio.NewReader(session.Conn), bufio.NewWriter(session.Conn))
	//Channel
	session.vChannel = 0
	session.vChannelControl = 1
	session.aChannel = 2
	session.aChannelControl = 3
	session.TransportType = TRANS_TYPE_TCP
	return session
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
