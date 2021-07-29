package RTSP

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestReadRequest(t *testing.T) {
	data := "DESCRIBE rtsp://admin:12345@192.0.1.100/Streaming/Channels/101 RTSP/1.0\r\nCSeq: 2\r\nAccept: application/sdp\r\nUser-Agent: NKPlayer\r\n\r\n"
	r := bytes.NewReader([]byte(data))
	req, err := ReadRequest(NewContext(), bufio.NewReader(r))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(req.String())
}
