package RTP

type Frame struct {
	SendType int
	DataLen  int
	Data     []byte
}
