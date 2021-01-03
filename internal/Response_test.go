package internal

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestReadResponse(t *testing.T) {
	data := `RTSP/1.0 401 Unauthorized
CSeq: 2
Www-Authenticate: Digest realm="Inphase Media Server", nonce="249c0a7ed", algorithm="MD5"

`
	rd := bufio.NewReader(strings.NewReader(data))
	resp,err := ReadResponse(rd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v",resp)
}