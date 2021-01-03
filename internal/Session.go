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

	Conn   *ConnRich
	ConnRW *bufio.ReadWriter

	vChannel        int
	vChannelControl int
	aChannel        int
	aChannelControl int
	Seq             int
	SdpInfos        []*SdpInfo

	Stoped bool
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
	return session
}
