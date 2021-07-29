package SDP

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	V_SDP = "video"
	A_SDP = "audio"
)

type SdpInfo struct {
	Codec       string
	TimeScale   int
	Control     string
	Rtpmap      int
	SpsPps      [][]byte
	PayloadType int
}

func ParseSdp(data string) (sdpInfo map[string]*SdpInfo, err error) {
	sdpInfo = make(map[string]*SdpInfo)
	var sdp *SdpInfo
	for _, line := range strings.Split(data, "\n") {
		line := strings.TrimSpace(line)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "m":
				fields := strings.SplitN(parts[1], " ", 4)
				if len(fields) == 4 {
					switch fields[0] {
					case V_SDP, A_SDP:
						sdp = new(SdpInfo)
						sdpInfo[fields[0]] = sdp
						sdp.PayloadType, _ = strconv.Atoi(fields[3])
					}
				}
			case "a":
				if sdp != nil {
					fields := strings.Split(parts[1], " ")
					for i := 0; i < len(fields); i++ {
						pFields := strings.SplitN(fields[i], ":", 2)
						if len(pFields) == 2 {
							switch pFields[0] {
							case "rtpmap":
								sdp.Rtpmap, _ = strconv.Atoi(pFields[1])
							case "control":
								sdp.Control = pFields[1]
							}
						}
						pFields = strings.SplitN(fields[i], "/", 2)
						if len(pFields) == 2 {
							switch pFields[0] {
							case "H264":
								sdp.Codec = "h264"
							case "PCMU":
								sdp.Codec = "pcm"
							}
							sdp.TimeScale, _ = strconv.Atoi(pFields[1])
						}
						for _, item := range strings.Split(fields[i], ";") {
							if mFields := strings.SplitN(item, "=", 2); len(mFields) == 2 {
								switch mFields[0] {
								case "sprop-parameter-sets":
									spspps := strings.Split(mFields[1], ",")
									for _, spspp := range spspps {
										info, _ := base64.StdEncoding.DecodeString(spspp)
										sdp.SpsPps = append(sdp.SpsPps, info)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if len(sdpInfo) == 0 {
		err = fmt.Errorf("Not Found Media Info")
	}
	return
}
