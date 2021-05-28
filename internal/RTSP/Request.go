package RTSP

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	RTSP_VERSION = "RTSP/1.0"
)

const (
	OPTIONS = "OPTIONS"

	DESCRIBE = "DESCRIBE"

	ANNOUNCE = "ANNOUNCE"

	SETUP = "SETUP"

	PLAY = "PLAY"

	PAUSE = "PAUSE"

	RECORD = "RECORD"

	REDIRECT = "REDIRECT"

	TEARDOWN = "TEARDOWN"
)

const (
	ContentLength    = "Content-Length"
	UserAgent        = "User-Agent"
	Authorization    = "Authorization"
	SessionID        = "Session"
	WWW_Authenticate = "WWW-Authenticate"
	Accept           = "Accept"
	Transport        = "Transport"
	Range            = "Range"
	CSeq             = "CSeq"
	Public           = "Public"
)

type Request struct {
	Method  string
	URL     string
	Version string
	Header  map[string]string
	Body    string
}

func ReadRequest(r *bufio.Reader) (req Request, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return
	}
	//Request-Line
	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("Request Line Format Error")
	}
	req.Method = parts[0]
	req.URL = parts[1]
	req.Version = parts[2]
	for {
		line, err = r.ReadString('\n')
		if err != nil {
			err = errors.Wrap(err, "Read Request Header Error")
			return
		}
		if len(strings.TrimSpace(line)) == 0 {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if req.Header == nil {
			req.Header = make(map[string]string)
		}
		if len(parts) != 2 {
			err = errors.New("Request Header Format Error")
			return
		}
		req.Header[parts[0]] = strings.TrimSpace(parts[1])
	}
	contentLengthStr, ok := req.Header[ContentLength]
	if ok {
		var contentLength int
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			err = errors.Wrap(err, "Request Header ContentLength Error")
			return
		}
		if contentLength > 0 {
			var data []byte
			data, err = r.Peek(contentLength)
			if err != nil {
				return
			}
			req.Body = string(data)
		}
	}
	return
}

func (r *Request) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL, r.Version))
	for key, val := range r.Header {
		str.WriteString(fmt.Sprintf("%s: %s\r\n", key, val))
	}
	str.WriteString("\r\n")
	str.WriteString(r.Body)
	return str.String()
}

func (r *Request) GetLength() int {
	v, err := strconv.ParseInt(r.Header[ContentLength], 10, 64)
	if err != nil {
		return 0
	} else {
		return int(v)
	}
}
