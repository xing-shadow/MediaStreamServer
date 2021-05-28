package RTSP

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Response struct {
	Version    string
	StatusCode int
	Status     string
	Header     map[string]string
	Body       string
}

func ReadResponse(r *bufio.Reader) (resp Response, err error) {
	respLine, err := r.ReadString('\n')
	if err != nil {
		return resp, err
	}
	parts := strings.SplitN(strings.TrimSpace(respLine), " ", 3)
	if len(parts) != 3 {
		err = errors.New("Read Rtsp Request Format Error")
		return
	}
	//Response-Line
	resp.Version = parts[0]
	resp.StatusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		err = errors.Wrap(err, "Rtsp Response Status Format Error")
		return
	}
	resp.Status = parts[2]
	//Response-Header
	for {
		var line string
		line, err = r.ReadString('\n')
		if err != nil {
			return resp, err
		}
		if len(strings.TrimSpace(line)) == 0 {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			err = errors.New("Rtsp Response Header Format Error")
			return
		} else {
			if resp.Header == nil {
				resp.Header = make(map[string]string)
			}
			resp.Header[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	//Response-Body
	contentLengthStr, ok := resp.Header[ContentLength]
	if ok {

		var contentLength int
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			err = errors.Wrap(err, "Rtsp Response Content-Length Foarmat Error")
			return
		}
		if contentLength > 0 {
			var data []byte
			data, err = r.Peek(contentLength)
			if err != nil {
				return
			}
			resp.Body = string(data)
		}
	}
	return
}

func GenerateResponse(code int, desc string, header map[string]string, body string) (resp Response) {
	resp.StatusCode = code
	resp.Status = desc
	resp.Header = header
	resp.Body = body
	return
}

func (r *Response) String() string {
	str := fmt.Sprintf("%s %d %s\r\n", r.Version, r.StatusCode, r.Status)
	for key, value := range r.Header {
		str += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	str += "\r\n"
	str += r.Body
	return str
}
