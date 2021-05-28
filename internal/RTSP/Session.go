package RTSP

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"regexp"
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
	SessionID   string
	Seq         string
	nonce       string
	realm       string
	ChannelCode string

	SessionType SessionType

	Server *RtspServer

	Conn       *ConnRich
	ConnRW     *bufio.ReadWriter
	ConnRwLock sync.Mutex

	sdpInfo         map[string]*SDP.SdpInfo
	sdpRaw          string
	vChannel        int
	vChannelControl int
	aChannel        int
	aChannelControl int

	RtpHandleFunc  []func(data []byte)
	StopHandleFunc []func()

	Stoped bool
	Exit   chan struct{}
}

func NewSession(conn net.Conn, srv *RtspServer) *Session {
	s := &Session{
		SessionID: fmt.Sprintf("%d", Snowflake.GenerateId()),
		realm:     "xing-shadow",
		Server:    srv,
		Conn:      NewConnRich(conn, time.Second*10),
		Stoped:    false,
		Exit:      make(chan struct{}),
	}
	s.ConnRW = bufio.NewReadWriter(bufio.NewReader(s.Conn), bufio.NewWriter(s.Conn))
	return s
}

func (s *Session) Start() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		s.Stop()
	}()
	for s.Stoped {
		buf1, err := s.ConnRW.ReadByte()
		if err != nil {
			Logger.GetLogger().Error("Read Connection:"+err.Error(), zap.String("ChannelCode", s.ChannelCode))
			return
		}
		if buf1 == MagicChar {

		} else {
			s.ConnRW.UnreadByte()
			req, err := ReadRequest(s.ConnRW.Reader)
			if err != nil {
				Logger.GetLogger().Error("Read Request Fail:"+err.Error(), zap.String("ChannelCode", s.ChannelCode))
				return
			}
			if err := s.HandleRtspRequest(req); err != nil {
				Logger.GetLogger().Error("Handle Request Fail:"+err.Error(), zap.String("ChannelCode", s.ChannelCode))
				return
			}
		}
	}
}

func (s *Session) HandleRtspRequest(req Request) (err error) {
	s.Seq = req.Header[CSeq]
	err = s.AuthRequest(req)
	if err != nil {
		return err
	}
	if req.Method != OPTIONS {
		if ok := s.checkAuth(req, req.Method); !ok {
			return
		}
	}
	switch req.Method {
	case OPTIONS:
		err = s.Options(req)
	case DESCRIBE:
	case ANNOUNCE:
		err = s.ANNOUNCE(req)
	case SETUP:
	case PLAY:
	case PAUSE:
	case TEARDOWN:

	}
	return
}

func (s *Session) AuthRequest(req Request) (err error) {
	defer func() {
		if err != nil {
			header := make(map[string]string)
			header[SessionID] = s.SessionID
			header[CSeq] = s.Seq
			resp := GenerateResponse(400, "Invalid Request", header, "")
			s.WriteResponse(resp)
		}
	}()
	var _u *url.URL
	_u, err = url.Parse(req.URL)
	if err != nil {
		return errors.New("Request Line Url Error")
	}
	var info ReqInfo
	info, err = getUrlInfo(_u.Path)
	if err != nil {
		return errors.New("Not Found Channel Code")
	}
	s.ChannelCode = info.ChannelCode
	return nil
}

func (s *Session) Options(req Request) error {
	header := make(map[string]string)
	header[SessionID] = s.SessionID
	header[CSeq] = s.Seq
	header[Public] = "DESCRIBE,PLAY,SETUP,TEARDOWN,ANNOUNCE"
	resp := GenerateResponse(200, "OK", header, "")
	s.WriteResponse(resp)
	return nil
}

func (s *Session) ANNOUNCE(req Request) (err error) {
	s.SessionType = SESION_TYPE_PUSHER
	s.sdpInfo, err = SDP.ParseSdp(req.Body)
	if err != nil {
		return err
	}
	NewPusher(s)
	header := make(map[string]string)
	header[SessionID] = s.SessionID
	header[CSeq] = s.Seq
	resp := GenerateResponse(200, "OK", header, "")
	s.WriteResponse(resp)
	return
}

func (s *Session) Setup(req Request) (err error) {
	return
}

func (s *Session) checkAuth(req Request, method string) bool {
	authLien, ok := req.Header[Authorization]
	if !ok {
		header := make(map[string]string)
		header[CSeq] = s.Seq
		header[SessionID] = s.SessionID
		s.nonce = fmt.Sprintf("%x", md5.Sum([]byte(strconv.FormatInt(Snowflake.GenerateId(), 10))))
		header[WWW_Authenticate] = fmt.Sprintf(`Digest realm="%s", nonce="%s", algorithm="MD5"`, s.realm, s.nonce)
		resp := GenerateResponse(401, "Unauthorized", header, "")
		s.WriteResponse(resp)
		return false
	} else {
		authFlag := s.digestAuth(authLien, method)
		if !authFlag {
			header := make(map[string]string)
			header[CSeq] = s.Seq
			header[SessionID] = s.SessionID
			resp := GenerateResponse(403, "Forbidden", header, "")
			s.WriteResponse(resp)
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

type ReqInfo struct {
	ChannelCode string
}

//rtsp://admin:admin@host/ChannelCode=xxx
func getUrlInfo(u string) (r ReqInfo, err error) {
	parts := strings.Split(u, "/")
	if len(parts) == 2 {
		r.ChannelCode = strings.TrimPrefix(parts[1], "ChannelCode=")
	} else {
		err = errors.New("NotFound")
	}
	return
}

func (s *Session) WriteResponse(resp Response) {
	s.ConnRwLock.Lock()
	s.ConnRW.WriteString(resp.String())
	s.ConnRW.Flush()
	s.ConnRwLock.Unlock()
}

func (s *Session) Stop() {
	if s.Stoped {
		return
	}
	s.Stoped = true
	close(s.Exit)
	for _, f := range s.StopHandleFunc {
		f()
	}
	if s.Conn != nil {
		s.Conn.conn.Close()
		s.Conn = nil
	}

}
