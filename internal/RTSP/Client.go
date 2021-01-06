package RTSP

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/internal/RTP"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"go.uber.org/zap"
	"io"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RtspClient struct {
	Sess     *Session
	authLine string
}

func NewRespClient(addr string) *RtspClient {
	return &RtspClient{
		Sess: &Session{Url: addr},
	}
}

func (c *RtspClient) StartPlayRealStream() (err error) {
	rtspAddr, err := url.Parse(c.Sess.Url)
	if err != nil {
		return err
	}
	var conn net.Conn
	if rtspAddr.Port() == "" {
		rtspAddr.Host = fmt.Sprintf("%s:%d", rtspAddr.Host, 554)
		c.Sess.Url = rtspAddr.String()
	}
	conn, err = net.Dial("tcp", rtspAddr.Host)
	c.Sess.NewRtspClientSession(conn, "")
	err = c.startRequestRealStream()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (c *RtspClient) startRequestRealStream() error {
	_, err := c.options()
	if err != nil {
		return err
	}
	resp, err := c.describe()
	if err != nil {
		return err
	}
	//TODO parse sdp
	sdpInfos := ParseSdp(resp.Body)
	if c.Sess.SdpInfos == nil {
		c.Sess.SdpInfos = make(map[string]*SdpInfo)
	}
	c.Sess.SdpInfos = sdpInfos
	resp, err = c.setup()
	if err != nil {
		return err
	}
	resp, err = c.play()
	if err != nil {
		return err
	}
	go c.streamReceiverStream()
	return nil
}

func (c *RtspClient) streamReceiverStream() {
	defer func() {
		c.Stop()
	}()
	logger := Logger.GetLogger()
	buf1 := make([]byte, 1)
	buf2 := make([]byte, 2)
	lastTime := time.Now()
	for !c.Sess.Stoped {
		if time.Now().Unix()-lastTime.Unix() > 30 {
			lastTime = time.Now()
			_, err := c.RequestNoResp(OPTIONS, nil)
			if err != nil {
				logger.Error("send KeepAlive info fail: "+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
				return
			}
		}
		if _, err := io.ReadFull(c.Sess.ConnRW, buf1); err != nil {
			logger.Error("read err:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
		}
		if buf1[0] == 0x24 {
			if _, err := io.ReadFull(c.Sess.ConnRW, buf1); err != nil {
				logger.Error("read err:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
				return
			}
			channel := buf1[0]
			if _, err := io.ReadFull(c.Sess.ConnRW, buf2); err != nil {
				logger.Error("read err:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
				return
			}
			rtpLen := binary.BigEndian.Uint16(buf2)
			if rtpLen > 65535 {
				logger.Error("read bad rtp pkg len gte 65535 bytes", zap.String("rtsp_addr", c.Sess.Url))
				return
			}
			data := make([]byte, rtpLen)
			_, err := io.ReadFull(c.Sess.ConnRW, data)
			if err != nil {
				logger.Error("read rtp data fail:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
				return
			}
			switch int(channel) {
			case c.Sess.vChannel:
				rtpPack, err := RTP.ParseRTPPack(data)
				if err != nil {
					logger.Error("parse rtp pack fail:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
					return
				}
				fmt.Println(rtpPack.Mark, rtpPack.PayloadType, rtpPack.Ts, rtpPack.SSRC)
				//case c.Sess.aChannel:
				//	fmt.Println("Receiver audio len: ", rtpLen)
			}
		} else {
			_ = c.Sess.ConnRW.UnreadByte()
			_, err := ReadResponse(c.Sess.ConnRW.Reader)
			if err != nil {
				logger.Error("receiver KeepAlive fail:"+err.Error(), zap.String("rtsp_addr", c.Sess.Url))
				return
			}
		}
	}
}

func (c *RtspClient) Stop() {
	if c.Sess.Stoped {
		return
	}
	c.Sess.Stop()
}

func (c *RtspClient) options() (resp Response, err error) {
	header := make(map[string]string)
	return c.Request(OPTIONS, header)
}

func (c *RtspClient) describe() (resp Response, err error) {
	header := make(map[string]string)
	header[Accept] = "application/sdp"
	return c.Request(DESCRIBE, header)
}

func (c *RtspClient) setup() (resp Response, err error) {
	var flag bool
	for _, sdpInfo := range c.Sess.SdpInfos {
		switch sdpInfo.AVType {
		case "video":
			flag = true
			c.Sess.vCodec = sdpInfo.Codec
			c.Sess.vControl = sdpInfo.Control
			var _url = ""
			if strings.Index(c.Sess.vControl, "rtsp://") == 0 {
				_url = c.Sess.vControl
			} else {
				_url = strings.TrimRight(c.Sess.Url, "/") + "/" + strings.TrimLeft(c.Sess.vControl, "/")
			}
			headers := make(map[string]string)
			if c.Sess.SessionID != "" {
				headers[SessionID] = c.Sess.SessionID
			}
			if c.Sess.TransportType == TRANS_TYPE_TCP {
				headers[Transport] = fmt.Sprintf("RTP/AVP/TCP;unicast;interleaved=%d-%d", c.Sess.vChannel, c.Sess.vChannelControl)
				var l *url.URL
				l, err = url.Parse(_url)
				if err != nil {
					return
				}
				l.User = nil
				resp, err = c.RequestWithPath(SETUP, l.String(), headers, true)
				if err != nil {
					return
				}
			}
		case "audio":
			flag = true
			c.Sess.aCodec = sdpInfo.Codec
			c.Sess.aControl = sdpInfo.Control
			var _url = ""
			if strings.Index(c.Sess.vControl, "rtsp://") == 0 {
				_url = c.Sess.vControl
			} else {
				_url = strings.TrimRight(c.Sess.Url, "/") + "/" + strings.TrimLeft(c.Sess.aControl, "/")
			}
			headers := make(map[string]string)
			if c.Sess.SessionID != "" {
				headers[SessionID] = c.Sess.SessionID
			}
			if c.Sess.TransportType == TRANS_TYPE_TCP {
				headers[Transport] = fmt.Sprintf("RTP/AVP/TCP;unicast;interleaved=%d-%d", c.Sess.aChannel, c.Sess.aChannelControl)
				var l *url.URL
				l, err = url.Parse(_url)
				if err != nil {
					return
				}
				l.User = nil
				resp, err = c.RequestWithPath(SETUP, l.String(), headers, true)
				if err != nil {
					return
				}
			}
		}
	}
	if flag {
		return
	} else {
		err = fmt.Errorf("not found video or audio sdp")
		return
	}
}

func (c *RtspClient) play() (resp Response, err error) {
	headers := make(map[string]string)
	headers[Range] = "npt=0.000-"
	return c.Request(PLAY, headers)
}

func (c *RtspClient) Request(method string, headers map[string]string) (resp Response, err error) {
	l, err := url.Parse(c.Sess.Url)
	if err != nil {
		return
	}
	l.User = nil
	return c.RequestWithPath(method, l.String(), headers, true)
}
func (c *RtspClient) RequestNoResp(method string, headers map[string]string) (resp Response, err error) {
	l, err := url.Parse(c.Sess.Url)
	if err != nil {
		return
	}
	l.User = nil
	return c.RequestWithPath(method, l.String(), headers, false)
}

func (c *RtspClient) RequestWithPath(method string, url string, headers map[string]string, needResp bool) (Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	logger := Logger.GetLogger()
	if c.Sess.Agent != "" {
		headers[UserAgent] = c.Sess.Agent
	}
	if len(headers[Authorization]) == 0 {
		if len(c.authLine) != 0 {
			authorization, err := Digest(method, c.authLine, c.Sess.Url)
			if err != nil {
				return Response{}, err
			}
			headers[Authorization] = authorization
		}
	}
	if len(c.Sess.SessionID) > 0 {
		headers[SessionID] = c.Sess.SessionID
	}
	c.Sess.Seq++
	headers[CSeq] = strconv.Itoa(c.Sess.Seq)
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("%s %s %s\r\n", method, url, RTSP_VERSION))
	for key, val := range headers {
		buff.WriteString(fmt.Sprintf("%s: %s\r\n", key, val))
	}
	buff.WriteString("\r\n")
	_, err := c.Sess.ConnRW.Write(buff.Bytes())
	if err != nil {
		logger.Error("write rtsp request info fail:" + err.Error())
		return Response{}, err
	}
	_ = c.Sess.ConnRW.Flush()
	if !needResp {
		return Response{}, nil
	}
	resp, err := ReadResponse(c.Sess.ConnRW.Reader)
	fmt.Println(resp)
	if err != nil {
		return Response{}, err
	}
	if val, ok := resp.Header[SessionID]; ok {
		c.Sess.SessionID = val
	}
	if err := c.checkAuth(method, resp); err != nil {
		return Response{}, nil
	} else {
		return resp, err
	}
}

func (c *RtspClient) checkAuth(method string, resp Response) error {
	if resp.StatusCode == 401 {
		val, ok := resp.Header[WWW_Authenticate]
		if !ok {
			return fmt.Errorf("status equeal 401 ,but not found www-authenticate")
		}
		c.authLine = val
		return nil
	} else {
		return nil
	}
}

func Digest(method string, authLine string, _url string) (authResp string, err error) {
	l, err := url.Parse(_url)
	if err != nil {
		return
	}
	realm := ""
	nonce := ""
	realmRex := regexp.MustCompile(`realm="(.*?)"`)
	result1 := realmRex.FindStringSubmatch(authLine)
	nonceRx := regexp.MustCompile(`nonce="(.*?)"`)
	result2 := nonceRx.FindStringSubmatch(authLine)
	if len(result1) == 2 {
		realm = result1[1]
	} else {
		err = fmt.Errorf("authline not found realm")
		return
	}
	if len(result2) == 2 {
		nonce = result2[1]
	} else {
		err = fmt.Errorf("authline not found nonce")
		return
	}
	username := l.User.Username()
	password, _ := l.User.Password()
	l.User = nil
	if l.Port() == "" {
		l.Host = fmt.Sprintf("%s:%d", l.Host, 554)
	}
	md5UserRealmPwd := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password))))
	md5MethodURL := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", method, l.String()))))
	response := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", md5UserRealmPwd, nonce, md5MethodURL))))
	authResp = fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"", username, realm, nonce, l.String(), response)
	return
}
