package RTSP

import (
	"fmt"
	"net"
	"sync"

	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
)

type RtspServer struct {
	Port     int
	listener net.Listener

	Server      *RtspServer
	connChannel chan net.Conn

	PushManager *PusherManager

	Exit   chan struct{}
	Closed bool
}

func NewRtspServer(port int) *RtspServer {
	rtspServer := &RtspServer{
		Port:        port,
		PushManager: NewPusherManager(),
	}
	return rtspServer
}

func (s *RtspServer) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	s.connChannel = make(chan net.Conn, 1)
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
		s := NewSession(conn, s)
		ctx := NewContext()
		go s.start(ctx)
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

func (pThis *PusherManager) remove(pusher *Pusher) {
	pThis.pusherLock.Lock()
	delete(pThis.pusher, pusher.Id)
	pThis.pusherLock.Unlock()
}
