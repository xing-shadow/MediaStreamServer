package RichConn

import (
	"net"
	"time"
)

type ConnRich struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Conn         net.Conn
}

func NewConnRich(conn net.Conn, readTimeout time.Duration, writeTimeout time.Duration) *ConnRich {
	return &ConnRich{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Conn:         conn,
	}
}

func (c *ConnRich) Write(p []byte) (n int, err error) {
	if c.WriteTimeout > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	} else {
		var t time.Time
		_ = c.Conn.SetWriteDeadline(t)
	}
	return c.Conn.Write(p)
}

func (c *ConnRich) Read(p []byte) (n int, err error) {
	if c.ReadTimeout > 0 {
		_ = c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	} else {
		var t time.Time
		_ = c.Conn.SetReadDeadline(t)
	}
	return c.Conn.Read(p)
}
