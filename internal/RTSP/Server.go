package RTSP

import "sync"

type RtspServer struct {
	pusherLock *sync.RWMutex
	pusher     map[string]*Pusher
}

func NewRtspServer() *RtspServer {
	rtspServer := &RtspServer{
		pusherLock: new(sync.RWMutex),
		pusher:     make(map[string]*Pusher),
	}
	return rtspServer
}
