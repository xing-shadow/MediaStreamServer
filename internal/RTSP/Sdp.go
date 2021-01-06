package RTSP

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
)

type SdpInfo struct {
	AVType             string
	Codec              string
	TimeScale          int
	Control            string
	Rtpmap             int
	Config             []byte
	SpropParameterSets [][]byte
	PayloadType        int
	SizeLength         int
	indexLength        int
}

func ParseSdp(SdpRaw string) map[string]*SdpInfo {
	sdpMap := make(map[string]*SdpInfo)
	var sdpInfo *SdpInfo
	for _, line := range strings.Split(SdpRaw, "\n") {
		line = strings.TrimSpace(line)
		typeval := strings.SplitN(line, "=", 2)
		if len(typeval) == 2 {
			fields := strings.SplitN(typeval[1], " ", 2)
			switch typeval[0] {
			case "m":
				switch fields[0] {
				case "video":
					sdpInfo = &SdpInfo{
						AVType: fields[0],
					}
					sdpMap["video"] = sdpInfo
					mfields := strings.Split(fields[1], " ")
					if len(mfields) >= 3 {
						sdpInfo.PayloadType, _ = strconv.Atoi(mfields[2])
					}
				case "audio":
					sdpInfo = &SdpInfo{
						AVType: fields[0],
					}
					sdpMap["audio"] = sdpInfo
					mfields := strings.Split(fields[1], " ")
					if len(mfields) >= 3 {
						sdpInfo.PayloadType, _ = strconv.Atoi(mfields[2])
					}
				}
			case "a":
				if sdpInfo != nil {
					for _, field := range fields {
						mfields := strings.SplitN(field, ":", 2)
						if len(mfields) >= 2 {
							switch mfields[0] {
							case "control":
								sdpInfo.Control = mfields[1]
							case "rtpmap":
								sdpInfo.Rtpmap, _ = strconv.Atoi(mfields[1])
							}
						}
						mfields = strings.Split(field, "/")
						if len(mfields) >= 2 {
							switch mfields[0] {
							case "MPEG4-GENERIC":
								sdpInfo.Codec = "aac"
							case "H264":
								sdpInfo.Codec = "h264"
							case "H265":
								sdpInfo.Codec = "h265"
							}
							if i, err := strconv.Atoi(mfields[1]); err == nil {
								sdpInfo.TimeScale = i
							}
						}
						mfields = strings.Split(field, ";")
						if len(mfields) > 1 {
							for _, mfield := range mfields {
								keyval := strings.SplitN(mfield, "=", 2)
								if len(keyval) == 2 {
									switch keyval[0] {
									case "config":
										sdpInfo.Config, _ = hex.DecodeString(keyval[1])
									case "sizelength":
										sdpInfo.SizeLength, _ = strconv.Atoi(keyval[1])
									case "indexlength":
										sdpInfo.indexLength, _ = strconv.Atoi(keyval[1])
									case "sprop-parameter-sets":
										ppsSps := strings.SplitN(keyval[1], ",", 2)
										for _, item := range ppsSps {
											val, _ := base64.StdEncoding.DecodeString(item)
											sdpInfo.SpropParameterSets = append(sdpInfo.SpropParameterSets, val)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return sdpMap
}
