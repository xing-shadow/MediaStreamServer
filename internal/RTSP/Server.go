package RTSP

import (
	"fmt"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Logger"
	"git.hub.com/wangyl/MediaSreamServer/pkg/Settings"
	"net"
)

type Option struct {
	Cfg Settings.RtspServer
}

func (o Option) fixme() {
	if o.Cfg.RtspPort <= 0 {
		o.Cfg.RtspPort = 554
	}
	if o.Cfg.ReadTimeout < 0 {
		o.Cfg.ReadTimeout = 10
	}
	if o.Cfg.WriteTimeout < 0 {
		o.Cfg.WriteTimeout = 10
	}
}

type RtspServer struct {
	opt         Option
	listener    net.Listener
	PushManager *PusherManager
	Exit        chan struct{}
	Closed      bool
}

func NewRtspServer(opt Option) *RtspServer {
	opt.fixme()
	rtspServer := &RtspServer{
		opt:         opt,
		PushManager: NewPusherManager(),
	}
	return rtspServer
}

func (s *RtspServer) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opt.Cfg.RtspPort))
	if err != nil {
		return err
	}
	s.Exit = make(chan struct{})
	s.listener = listener
	go s.handleConn()
	return nil
}

func (s *RtspServer) handleConn() {
	for !s.Closed {
		conn, err := s.listener.Accept()
		if err != nil {
			Logger.GetLogger().Error("accept err: " + err.Error())
			continue
		}
		ctx := NewContext()
		s := NewSession(ctx, conn, s)
		go s.start()
	}
}

func (s *RtspServer) Stop() {
	if s.Closed {
		return
	}
	close(s.Exit)
}
