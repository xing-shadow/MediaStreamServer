package internal

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Response struct {
	Version    string
	StatusCode int
	Status     string
	Header     map[string]string
	Body       string
}

func ReadResponse(rd *bufio.Reader) (Response,error) {
	var resp = Response{
		Header: make(map[string]string),
	}
	respLine,err := rd.ReadString('\n')
	if err != nil {
		return resp, err
	}
	parts := strings.Split(respLine," ")
	if len(parts) != 3 {
		return Response{},nil
	}
	//Response-Line
	resp.Version = parts[0]
	resp.StatusCode,err = strconv.Atoi(parts[1])
	if err != nil {
		return resp, err
	}
	resp.Status = parts[2]
	//Response-Header
	for {
		line,err := rd.ReadString('\n')
		if err != nil {
			return resp,err
		}
		lineTrimSpace := strings.TrimSpace(line)
		if  len(lineTrimSpace) == 0 {
			break
		}
		parts := strings.SplitN(lineTrimSpace,":",2)
		if len(parts) < 2 {
			return resp,errors.New("parse resp header invalid")
		}else {
			resp.Header[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	//Response-Body
	contentLength,_ := strconv.Atoi(resp.Header[ContentLength])
	if contentLength > 0 {
		body,err := rd.Peek(contentLength)
		if err != nil {
			return resp,err
		}
		resp.Body = string(body)
	}
	return resp,nil
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