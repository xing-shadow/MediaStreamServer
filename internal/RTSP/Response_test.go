package RTSP

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestReadResponse(t *testing.T) {
	data := "RTSP/1.0 200 OK\r\nCSeq: 3\r\nSession:1389957320;timeout=60//服务器回应的会话标识符和超时时间\r\nTransport: RTP/AVP;unicast;client_port=1094-1095;server_port=12028-1202\r\n\r\n"
	r := bytes.NewReader([]byte(data))
	resp, err := ReadResponse(bufio.NewReader(r))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp.String())
}
