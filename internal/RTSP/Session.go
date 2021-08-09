package RTSP

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RichConn"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/SDP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Snowflake"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Session struct {
	ctx         *Context
	sessionID   string
	seq         string
	nonce       string
	realm       string
	channelCode string

	SessionType SessionType

	Server *RtspServer

	richConn   *RichConn.ConnRich
	connRW     *bufio.ReadWriter
	connRwLock sync.Mutex

	sdpInfo         map[string]*SDP.SdpInfo
	sdpRaw          string
	vControl        string
	aControl        string
	vChannel        int
	vChannelControl int
	aChannel        int
	aChannelControl int

	rtpHandleFunc  []func(frame RTP.Frame)
	stopHandleFunc []func()

	stoped bool
}

func NewSession(ctx *Context, conn net.Conn, srv *RtspServer) *Session {
	s := &Session{
		ctx:       ctx,
		sessionID: fmt.Sprintf("%d", Snowflake.GenerateId()),
		realm:     "xing-shadow",
		Server:    srv,
		richConn:  RichConn.NewConnRich(conn, time.Second*time.Duration(srv.opt.Cfg.ReadTimeout), time.Second*time.Duration(srv.opt.Cfg.WriteTimeout)),
		stoped:    false,
	}
	s.connRW = bufio.NewReadWriter(bufio.NewReader(s.richConn), bufio.NewWriter(s.richConn))
	return s
}

func (s *Session) start() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1638)
			buf = buf[:runtime.Stack(buf, false)]
			pl := fmt.Sprintf("Panic: %v\n%s\n", err, buf)
			fmt.Fprintf(os.Stderr, pl)
		}
		s.stop()
	}()
	for !s.stoped {
		magic, err := s.connRW.ReadByte()
		if err != nil {
			Logger.GetLogger().Error("Read Connection:"+err.Error(), zap.String("ChannelCode", s.channelCode))
			return
		}
		if magic == MagicChar {
			buf1, err := s.connRW.ReadByte()
			if err != nil {
				Logger.GetLogger().Error("Read Connection:"+err.Error(), zap.String("ChannelCode", s.channelCode))
				return
			}
			channel := int(buf1)
			buf2 := make([]byte, 2)
			if _, err := io.ReadFull(s.connRW, buf2); err != nil {
				Logger.GetLogger().Error("Read Connection:"+err.Error(), zap.String("ChannelCode", s.channelCode))
				return
			}
			rtpLen := binary.BigEndian.Uint16(buf2)
			if rtpLen > 65535 {
				Logger.GetLogger().Error("get rtp packet length more than 65535", zap.String("ChannelCode", s.channelCode))
				return
			}
			rtpData := make([]byte, rtpLen)
			if _, err := io.ReadFull(s.connRW, rtpData); err != nil {
				Logger.GetLogger().Error("Read Connection:"+err.Error(), zap.String("ChannelCode", s.channelCode))
				return
			}
			switch channel {
			case s.aChannel:
				var frame = RTP.Frame{
					SendType: RTP_TYPE_AUDIO,
					Data:     rtpData,
					DataLen:  int(rtpLen),
				}
				for _, f := range s.rtpHandleFunc {
					f(frame)
				}
			case s.aChannelControl:
				var frame = RTP.Frame{
					SendType: RTP_TYPE_AUDIOCONTROL,
					Data:     rtpData,
					DataLen:  int(rtpLen),
				}
				for _, f := range s.rtpHandleFunc {
					f(frame)
				}
			case s.vChannel:
				var frame = RTP.Frame{
					SendType: RTP_TYPE_VEDIO,
					Data:     rtpData,
					DataLen:  int(rtpLen),
				}
				for _, f := range s.rtpHandleFunc {
					f(frame)
				}
			case s.vChannelControl:
				var frame = RTP.Frame{
					SendType: RTP_TYPE_VIDEOCONTROL,
					Data:     rtpData,
					DataLen:  int(rtpLen),
				}
				for _, f := range s.rtpHandleFunc {
					f(frame)
				}
			}
		} else {
			err := s.connRW.UnreadByte()
			if err != nil {
				Logger.GetLogger().Error("UnreadByte Fail:"+err.Error(), zap.String("ChannelCode", s.channelCode))
				return
			}
			if err := s.HandleRtspRequest(); err != nil {
				Logger.GetLogger().Error("Handle Request Fail:"+err.Error(), zap.String("ChannelCode", s.channelCode))
				return
			}
		}
	}
}

func (s *Session) HandleRtspRequest() (err error) {
	defer func() {
		s.HandleRtspResponse()
	}()
	var req Request
	req, err = ReadRequest(s.ctx, s.connRW.Reader)
	if err != nil {
		header := make(map[string]string)
		header[SessionID] = s.sessionID
		header[CSeq] = s.seq
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
		return
	}
	s.ctx.req = req
	s.seq = req.Header[CSeq]
	err = s.AuthRequest()
	if err != nil {
		header := make(map[string]string)
		header[SessionID] = s.sessionID
		header[CSeq] = s.seq
		s.ctx.resp = GenerateResponse(400, "Invalid Request", header, "")
		return err
	}
	//if req.Method != OPTIONS {
	//	if ok := s.checkAuth(); !ok {
	//		return
	//	}
	//}
	switch req.Method {
	case OPTIONS:
		err = s.Options()
	case DESCRIBE:
		err = s.Describe()
	case ANNOUNCE:
		err = s.ANNOUNCE()
	case SETUP:
		err = s.Setup()
	case PLAY:
		err = s.Play()
	case RECORD:
		err = s.Record()
	case PAUSE:
	case TEARDOWN:
		err = s.Teardown()
	default:
		header := make(map[string]string)
		header[SessionID] = s.sessionID
		header[CSeq] = s.seq
		s.ctx.resp = GenerateResponse(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), header, "")
	}
	return
}

func (s *Session) HandleRtspResponse() {
	s.connRwLock.Lock()
	s.connRW.WriteString(s.ctx.resp.String())
	s.connRW.Flush()
	s.connRwLock.Unlock()
	if s.ctx.method == TEARDOWN {
		s.stop()
	}
	if s.ctx.resp.StatusCode != http.StatusOK && s.ctx.resp.StatusCode != http.StatusUnauthorized && s.ctx.resp.StatusCode != StatusCodeNotAccept {
		s.stop()
	}
}

//rtsp://admin:admin@host/ChannelCode=xxx
func (s *Session) AuthRequest() (err error) {
	parts := strings.Split(s.ctx.url.Path, "/")
	var channelCode string
	if len(parts) > 1 {
		data := strings.Split(parts[1], "=")
		if len(data) == 2 {
			channelCode = data[1]
		} else {
			err = errors.New("url format error")
			return
		}
	}
	if len(channelCode) == 0 {
		return errors.New("not found channelCode")
	}
	if len(s.channelCode) != 0 && s.channelCode != channelCode {
		return errors.New("channelCode mismatch")
	} else {
		s.channelCode = channelCode
	}
	return nil
}

func (s *Session) Options() error {
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	header[Public] = "DESCRIBE,PLAY,SETUP,TEARDOWN,ANNOUNCE"
	s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
	return nil
}

func (s *Session) Describe() (err error) {
	s.SessionType = SESSION_TYPE_PLAYER
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	pusher, isExit := s.Server.PushManager.pusherIsExit(s.channelCode)
	if !isExit {
		err = errors.New("Not Found Pusher")
		s.ctx.resp = GenerateResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), header, "")
		return
	}
	if mediaInfo, ok := pusher.s.sdpInfo["video"]; ok {
		s.vControl, err = getControl(mediaInfo)
	}
	if mediaInfo, ok := pusher.s.sdpInfo["audio"]; ok {
		s.aControl, err = getControl(mediaInfo)
	}
	NewPlayer(pusher, s)
	s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, pusher.getSdp())
	return
}

func (s *Session) ANNOUNCE() (err error) {
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	s.SessionType = SESION_TYPE_PUSHER
	s.sdpInfo, err = SDP.ParseSdp(s.ctx.req.Body)
	if err != nil {
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
		return err
	}
	s.sdpRaw = s.ctx.req.Body
	if _, isExit := NewPusher(s); isExit {
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
		return
	}
	if mediaInfo, ok := s.sdpInfo["video"]; ok {
		s.vControl, err = getControl(mediaInfo)
	}
	if mediaInfo, ok := s.sdpInfo["audio"]; ok {
		s.aControl, err = getControl(mediaInfo)
	}
	if err != nil {
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
	} else {
		s.ctx.resp = GenerateResponse(200, "OK", header, "")
	}
	return
}

func getControl(sdp *SDP.SdpInfo) (string, error) {
	if strings.Index(strings.ToLower(sdp.Control), "rtsp://") == 0 {
		controlUrl, err := url.Parse(sdp.Control)
		if err != nil {
			return "", err
		}
		return controlUrl.String(), nil
	} else {
		return sdp.Control, nil
	}
}

func (s *Session) Setup() (err error) {
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	ts, ok := s.ctx.req.Header[Transport]
	if !ok {
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
		return errors.New("setup not found transport")
	}
	parts := strings.Split(ts, "/TCP;")
	if len(parts) == 2 { //tcp发流
		if tcpMatch := TcpRegexp.FindStringSubmatch(parts[1]); tcpMatch != nil {
			setupPath := s.ctx.url.String()
			if setupPath == s.vControl || (strings.Contains(setupPath, s.vControl) && strings.LastIndex(setupPath, s.vControl) == len(setupPath)-len(s.vControl)) {
				s.vChannel, err = strconv.Atoi(tcpMatch[1])
				if err != nil {
					s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
					return
				}
				s.vChannelControl, err = strconv.Atoi(tcpMatch[3])
				if err != nil {
					s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
					return
				} else {
					header[Transport] = ts
					s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
					return
				}
			}
			if setupPath == s.aControl || (strings.Contains(setupPath, s.aControl) && strings.LastIndex(setupPath, s.aControl) == len(setupPath)-len(s.aControl)) {
				s.aChannel, err = strconv.Atoi(tcpMatch[1])
				if err != nil {
					s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
					return
				}
				s.aChannelControl, err = strconv.Atoi(tcpMatch[3])
				if err != nil {
					s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
					return
				} else {
					header[Transport] = ts
					s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
					return
				}
			}
		}
	} else { //不支持udp
		s.ctx.resp = GenerateResponse(StatusCodeNotAccept, "Unsupported Transport", header, "")
	}
	return
}

func (s *Session) Play() (err error) {
	s.richConn.ReadTimeout = 0
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	header[Range] = "npt=0.000-"
	if pusher, isExit := s.Server.PushManager.pusherIsExit(s.channelCode); isExit {
		if player, isExit := pusher.getPlayer(s.sessionID); isExit {
			go player.receiverFrame()
			s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
		} else {
			s.ctx.resp = GenerateResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), header, "")
		}
	} else {
		s.ctx.resp = GenerateResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), header, "")
	}
	return
}

func (s *Session) Record() (err error) {
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	if pusher, isExit := s.Server.PushManager.pusherIsExit(s.channelCode); isExit {
		go pusher.ReceiveRtp()
		s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
	} else {
		s.ctx.resp = GenerateResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), header, "")
		err = errors.New("not found pusher")
	}
	return
}

func (s *Session) Teardown() (err error) {
	header := make(map[string]string)
	header[SessionID] = s.sessionID
	header[CSeq] = s.seq
	s.ctx.resp = GenerateResponse(http.StatusOK, http.StatusText(http.StatusOK), header, "")
	return
}

func (s *Session) checkAuth() bool {
	authLien, ok := s.ctx.req.Header[Authorization]
	if !ok {
		header := make(map[string]string)
		header[CSeq] = s.seq
		header[SessionID] = s.sessionID
		s.nonce = fmt.Sprintf("%x", md5.Sum([]byte(strconv.FormatInt(Snowflake.GenerateId(), 10))))
		header[WWW_Authenticate] = fmt.Sprintf(`Digest realm="%s", nonce="%s", algorithm="MD5"`, s.realm, s.nonce)
		s.ctx.resp = GenerateResponse(401, "Unauthorized", header, "")
		return false
	} else {
		authFlag := s.digestAuth(authLien, s.ctx.method)
		if !authFlag {
			header := make(map[string]string)
			header[CSeq] = s.seq
			header[SessionID] = s.sessionID
			s.ctx.resp = GenerateResponse(403, "Forbidden", header, "")
		}
		return authFlag
	}
}

/*
	HA1=MD5(username:realm:password)
	HA2=MD5(method:uri)
	Response =MD5(HA1:nonce:HA2)
*/
func (s *Session) digestAuth(auth string, method string) bool {
	usernameReg := regexp.MustCompile(`username="(.*?)"`)
	uriReg := regexp.MustCompile(`uri="(.*?)"`)
	respReg := regexp.MustCompile(`response="(.*?)"`)
	var username string
	if parts := usernameReg.FindStringSubmatch(auth); len(parts) == 2 {
		username = parts[1]
	}
	var uri string
	if parts := uriReg.FindStringSubmatch(auth); len(parts) == 2 {
		uri = parts[1]
	}
	var resp string
	if parts := respReg.FindStringSubmatch(auth); len(parts) == 2 {
		resp = parts[1]
	}
	ha1 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, s.realm, "admin"))))
	ha2 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri))))
	authResp := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", ha1, s.nonce, ha2))))
	if authResp == resp {
		return true
	} else {
		return false
	}
}

func (s *Session) stop() {
	if s.stoped {
		return
	}
	s.stoped = true
	for _, f := range s.stopHandleFunc {
		f()
	}
	if s.richConn != nil {
		s.richConn.Conn.Close()
		s.richConn = nil
	}
}
