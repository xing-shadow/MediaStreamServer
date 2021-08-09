package RTSP

import (
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Settings"
	"net"
	"sync"

	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
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
