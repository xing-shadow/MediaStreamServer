package RTSP

import "regexp"

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

type SessionType int

const (
	SESSION_TYPE_PLAYER = iota
	SESION_TYPE_PUSHER
)

const MagicChar = 0x24

const StatusCodeNotAccept = 461 //

var TcpRegexp = regexp.MustCompile("interleaved=(\\d+)(-(\\d+))?")
