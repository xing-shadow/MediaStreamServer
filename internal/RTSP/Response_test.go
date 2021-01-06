package RTSP

import (
	"testing"
)

func TestReadResponse(t *testing.T) {
	var client = NewRespClient("rtsp://admin:admin123@171.221.244.37:33556")
	if err := client.StartPlayRealStream(); err != nil {
		t.Fatal(err)
	}
	select {}
}
