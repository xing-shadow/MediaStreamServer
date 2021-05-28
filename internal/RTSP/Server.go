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

func (pThis *PusherManager) AddPusher(pusher *Pusher) {
	pThis.pusherLock.RLock()
	old, ok := pThis.pusher[pusher.Id]
	pThis.pusherLock.RUnlock()
	if ok {
		old.playerMutex.Lock()
		for key, player := range old.player {
			pusher.player[key] = player
		}
		old.playerMutex.Unlock()
		old.Stop()
	}
	pThis.pusherLock.Lock()
	pThis.pusher[pusher.Id] = pusher
	pThis.pusherLock.Unlock()
}

func (pThis *PusherManager) Remove(pusher *Pusher) {
	pThis.pusherLock.Lock()
	delete(pThis.pusher, pusher.Id)
	pThis.pusherLock.Unlock()
}

func (s *RtspServer) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	s.connChannel = make(chan net.Conn, 1)
	s.Exit = make(chan struct{})
	s.listener = listener
	go func() {
		for {
			select {
			case conn, ok := <-s.connChannel:
				if !ok {
					break
				}
				s := NewSession(conn, s)
				go s.Start()
			case <-s.Exit:
				break
			}
		}
	}()
	return nil
}

func (s *RtspServer) Accept() {
	for !s.Closed {
		conn, err := s.listener.Accept()
		if err != nil {
			Logger.GetLogger().Error("accept err: " + err.Error())
			continue
		}
		s.connChannel <- conn
	}
}

func (s *RtspServer) Stop() {
	if s.Closed {
		return
	}
	close(s.Exit)
}
