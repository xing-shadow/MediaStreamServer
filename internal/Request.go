package internal

import (
	"fmt"
	"strconv"
	"strings"
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
	Require          = "Require"
	WWW_Authenticate = "WWW-Authenticate"
	Accept           = "Accept"
	Transport        = "Transport"
	Range            = "Range"
	CSeq             = "CSeq"
)

type Request struct {
	Method  string
	URL     string
	Version string
	Header  map[string]string
	Content string
	Body    string
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
