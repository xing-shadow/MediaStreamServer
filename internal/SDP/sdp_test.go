package SDP

import (
	"fmt"
	"testing"
)

func TestParseSdp(t *testing.T) {
	var data string
	data = `
	v=0//SDP 版本信息
	o=- 1109162014219182 1109162014219192 IN IP4 x.y.z.w
	s=Media Presentation
	e=NONE
	c=IN IP4 0.0.0.0
	t=0 0
	m=video 0 RTP/AVP 96
	a=rtpmap:96 H264/90000
	a=control:trackID=1
	a=fmtp:96 profile-level-id=4D0014;packetization-mode=0;sprop-parameter-sets=Z0LAH4iLUCgC3QgAADhAAAr8gBA=,aM44gA==
	m=audio 0 RTP/AVP 0
	a=rtpmap:0 PCMU/8000
	a=control:trackID=2
	a=Media_header:MEDIAINFO=494D4B48010100000400010010710110401F000000FA000000000000000000000000000000000000;
	a=appversion:1.0
`
	sdp, err := ParseSdp(data)
	if err != nil {
		t.Fatal(err)
	}
	for media, sdpInfo := range sdp {
		fmt.Println(media)
		fmt.Printf("\t%+v\n", *sdpInfo)
	}
}
