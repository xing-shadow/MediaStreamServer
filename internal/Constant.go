package internal

type TransType int

const (
	TRANS_TYPE_TCP = iota
	TRAMS_TYPE_UDP
)

type RTPType int

const (
	RTP_TYPE_AUDIO = iota
	RTP_TYPE_VEDIO
	RTP_TYPE_AUDIOCONTROL
	RTP_TYPE_VIDEOCONTROL
)

type SESSION_TYPE int

const (
	SESSION_TYPE_PLAYER = iota
	SESION_TYPE_PUSHER
)
