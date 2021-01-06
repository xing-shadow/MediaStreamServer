package RTSP

import (
	"net"
	"time"
)

type ConnRich struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	conn         net.Conn
}

func NewConnRich(conn net.Conn) *ConnRich {
	return &ConnRich{
		conn: conn,
	}
}

func (c *ConnRich) SetReadTimeout(n time.Duration) {
	c.ReadTimeout = n
}

func (c *ConnRich) SetWriteTimeout(n time.Duration) {
	c.WriteTimeout = n
}

func (c *ConnRich) Write(p []byte) (n int, err error) {
	if c.WriteTimeout > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	} else {
		var t time.Time
		_ = c.conn.SetWriteDeadline(t)
	}
	return c.conn.Write(p)
}

func (c *ConnRich) Read(p []byte) (n int, err error) {
	if c.ReadTimeout > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	} else {
		var t time.Time
		_ = c.conn.SetReadDeadline(t)
	}
	return c.conn.Read(p)
}
