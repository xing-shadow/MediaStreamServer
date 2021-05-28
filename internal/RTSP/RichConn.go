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

func NewConnRich(conn net.Conn, timeout time.Duration) *ConnRich {
	return &ConnRich{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		conn:         conn,
	}
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
