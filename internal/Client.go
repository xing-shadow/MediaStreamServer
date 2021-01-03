package internal

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"git.hub.com/wangyl/RTSP_AGREEMENT/pkg/Logger"
	"net"
	"net/url"
	"regexp"
	"strings"
)

type RtspClient struct {
	Sess *Session
	authorization string
}

func NewRespClient(addr string) *RtspClient {
	return &RtspClient{
		Sess: &Session{Url:addr},
	}
}

func (c *RtspClient) StartPlayRealStream() (err error) {
	rtspAddr,err := url.Parse(c.Sess.Url)
	if err != nil {
		return err
	}
	var conn net.Conn
	if rtspAddr.Port() == "" {
		rtspAddr.Host = fmt.Sprintf("%s:%d",rtspAddr.Host,554)
		c.Sess.Url = rtspAddr.String()
	}
	conn,err = net.Dial("tcp",rtspAddr.Host)
	c.Sess = NewRtspClientSession(conn,"")

}

func (c *RtspClient) startRequestRealStream() error {
	_,err := c.options()
	if err != nil {
		return err
	}
	resp,err := c.describe()
	if err != nil {
		return err
	}
	//TODO parse sdp
}

func (c *RtspClient) options() (resp Response,err error) {
	header := make(map[string]string)
	header[Require] = "implicit-play"
	return c.Request(OPTIONS,header)
}

func (c *RtspClient) describe() (resp Response,err error) {
	header := make(map[string]string)
	header[Accept] = "application/sdp"
	return c.Request(DESCRIBE,header)
}

func (c *RtspClient) Request(method string,headers map[string]string) (resp Response,err error) {
	l,err := url.Parse(c.Sess.Url)
	if err != nil {
		return
	}
	l.User = nil
	return c.RequestWithPath(method,l.String(),headers,true)
}

func (c *RtspClient) RequestWithPath(method string,url string,headers map[string]string,needResp bool) (Response,error) {
	logger := Logger.GetLogger()
	headers[UserAgent] = c.Sess.Agent
	if len(headers[Authorization]) == 0 {
		if len(c.authorization) != 0 {
			headers[Authorization] = c.authorization
		}
	}
	if len(c.Sess.SessionID) > 0 {
		headers[SessionID] = c.Sess.SessionID
	}
	c.Sess.Seq++
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("%s %s %s\r\n",method,url,RTSP_VERSION))
	for key, val := range headers {
		buff.WriteString(fmt.Sprintf("%s: %s\r\n",key,val))
	}
	buff.WriteString("\r\n")
	_,err := c.Sess.ConnRW.Write(buff.Bytes())
	if err != nil {
		logger.Error("write rtsp request info fail:" + err.Error())
		return Response{},err
	}
	if !needResp {
		return Response{},nil
	}
	resp,err := ReadResponse(c.Sess.ConnRW.Reader)
	if err != nil {
		return Response{},err
	}
	if val,ok := resp.Header[SessionID]; ok {
		c.Sess.SessionID = val
	}
	if err := c.checkAuth(method,resp);err != nil{
		return Response{},nil
	}else {
		return resp,err
	}
}

func (c *RtspClient) checkAuth(method string,resp Response) error {
	if resp.StatusCode == 401 {
		val,ok := resp.Header[WWW_Authenticate]
		if !ok {
			return fmt.Errorf("status equeal 401 ,but not found www-authenticate")
		}
		if strings.Index(val,"Digest") == 0 {
			authorization,err := Digest(method,val,c.Sess.Url)
			if err != nil {
				return err
			}
			c.authorization = authorization
			return nil
		}else {
			return fmt.Errorf("status equeal 401,but auth algorithm no support")
		}
	}else {
		return nil
	}
}

func Digest(method string,authLine string,_url string) (authResp string,err error) {
	l,err := url.Parse(_url)
	if err != nil {
		return
	}
	realm := ""
	nonce := ""
	realmRex := regexp.MustCompile(`realm="(.*?)"`)
	result1 := realmRex.FindStringSubmatch(authLine)
	nonceRx := regexp.MustCompile(`nonce=(.*?)`)
	result2 := nonceRx.FindStringSubmatch(authLine)
	if len(result1) == 2 {
		realm = result1[1]
	}else {
		err = fmt.Errorf("authline not found realm")
		return
	}
	if len(result2) == 2 {
		nonce = result2[1]
	}else {
		err = fmt.Errorf("authline not found nonce")
		return
	}
	username := l.User.Username()
	password,_ := l.User.Password()
	l.User = nil
	if l.Port() == "" {
		l.Host = fmt.Sprintf("%s:%d",l.Host,554)
	}
	md5UserRealmPwd := fmt.Sprintf("%x",md5.Sum([]byte(fmt.Sprintf("%s:%s:%s",username,realm,password))))
	md5MethodURL := fmt.Sprintf("%x",md5.Sum([]byte(fmt.Sprintf("%s:%s",method,l.String()))))
	response := fmt.Sprintf("%x",md5.Sum([]byte(fmt.Sprintf("%s:%s:%s",md5UserRealmPwd,nonce,md5MethodURL))))
	authLine = fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"",username,realm,nonce,l.String(),response)
	return
}