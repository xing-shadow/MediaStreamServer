package app

import (
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTSP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"net"
	"runtime"
)

type RtspService struct {
	Port        int
	listener    net.Listener
	Server      *RTSP.RtspServer
	connChannel chan net.Conn

	Stoped chan bool
	Closed bool
}

func (s *RtspService) Init(port int) error {
	s.Port = port
	s.Server = RTSP.NewRtspServer()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	} else {
		s.connChannel = make(chan net.Conn, runtime.NumCPU()*2)
		s.listener = listener
		return nil
	}
}

func (s *RtspService) StartWork() {
	for true {
		select {
		//TODO set conn
		//case conn, ok := <-s.connChannel:
		//	if !ok {
		//		break
		//	}

		case <-s.Stoped:
			break
		}
	}
	Logger.GetLogger().Info("Rtsp Service stop")
}

func (s *RtspService) Accept() {
	for !s.Closed {
		conn, err := s.listener.Accept()
		if err != nil {
			Logger.GetLogger().Error("accept err: " + err.Error())
			continue
		}
		s.connChannel <- conn
	}
}

func (s *RtspService) Stop() {

}
