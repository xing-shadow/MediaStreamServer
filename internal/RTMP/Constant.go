package RTMP

const (
	TcpBufferSize = 1024 * 1024 * 1
)

type MessageType int

//rtmp消息类型
const (
	SetChunkSizeMsgType        MessageType = 1
	AbortMessageMsgType        MessageType = 2
	AckMsgType                 MessageType = 3
	UserControlMessagesMsgType MessageType = 4
	WindowAckSizeMsgType       MessageType = 5
	SetPeerBandwidthMsgType    MessageType = 6
	AudioMsgType               MessageType = 8
	VideoMsgType               MessageType = 9
	CmdAMF3MsgType             MessageType = 17
	CmdAMF0MsgType             MessageType = 20
	DataAMF0MsgType            MessageType = 18
	DataAMF3MsgType            MessageType = 15
	ShareAMF0MsgType           MessageType = 19
	ShareAMF3MsgType           MessageType = 16
)

//命令类型消息
var (
	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

//用户控制消息
const (
	streamBegin      uint32 = 0
	streamEOF        uint32 = 1
	streamDry        uint32 = 2
	setBufferLen     uint32 = 3
	streamIsRecorded uint32 = 4
	pingRequest      uint32 = 6
	pingResponse     uint32 = 7
)

const (
	chunkSize       = 1024
	remoteChunkSize = 128
)
