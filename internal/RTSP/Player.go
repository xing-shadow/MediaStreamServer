package RTSP

type Player struct {
	s *Session
}

func (pThis *Player) Stop() {
	pThis.s.Stop()
}
