package RTMP

import (
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Settings"
	"net"
	"sync"
)

type Option struct {
	Cfg Settings.RtmpServer
}

func (o Option) fixme() {
	if o.Cfg.RtmpPort <= 0 {
		o.Cfg.RtmpPort = 1935
	}
	if o.Cfg.ReadTimeout < 0 {
		o.Cfg.ReadTimeout = 10
	}
	if o.Cfg.WriteTimeout < 0 {
		o.Cfg.WriteTimeout = 10
	}
}

type RtmpServer struct {
	opt         Option
	listener    net.Listener
	PushManager *PusherManager
	Exit        chan struct{}
	Closed      bool
}

func NewRtmpServer(opt Option) *RtmpServer {
	opt.fixme()
	rtspServer := &RtmpServer{
		opt:         opt,
		PushManager: NewPusherManager(),
	}
	return rtspServer
}

func (s *RtmpServer) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opt.Cfg.RtmpPort))
	if err != nil {
		return err
	}
	s.Exit = make(chan struct{})
	s.listener = listener
	go s.handleConn()
	return nil
}

func (s *RtmpServer) handleConn() {
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

func (s *RtmpServer) Stop() {
	if s.Closed {
		return
	}
	close(s.Exit)
}

type PusherManager struct {
	pusherLock *sync.RWMutex
	pusher     map[string]*Pusher
}

func NewPusherManager() *PusherManager {
	return &PusherManager{
		pusherLock: new(sync.RWMutex),
		pusher:     make(map[string]*Pusher),
	}
}

func (pThis *PusherManager) addPusher(pusher *Pusher) (old *Pusher, isExit bool) {
	pThis.pusherLock.Lock()
	defer pThis.pusherLock.Unlock()
	if old, isExit = pThis.pusher[pusher.Id]; isExit {
		return
	} else {
		pThis.pusher[pusher.Id] = pusher
		return pusher, false
	}
}

func (pThis *PusherManager) pusherIsExit(id string) (pusher *Pusher, isExit bool) {
	pThis.pusherLock.RLock()
	defer pThis.pusherLock.RUnlock()
	if pusher, isExit = pThis.pusher[id]; isExit {
		return
	} else {
		return nil, false
	}
}

func (pThis *PusherManager) removePusher(pusher *Pusher) {
	pThis.pusherLock.Lock()
	delete(pThis.pusher, pusher.Id)
	pThis.pusherLock.Unlock()
}
