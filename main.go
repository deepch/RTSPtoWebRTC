package main

import (
	// "encoding/binary"
	"fmt"
	// "log"

	rtsp "github.com/deepch/sample_rtsp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

var (
	VideoWidth  int
	VideoHeight int
)

func main() {

	go StartHTTPServer()

	url := "rtsp://admin:wj123456@10.1.59.14:554/h264/ch1/main/av_stream"

	Client := rtsp.RtspClientNew()
	Client.Debug = false

	if err := Client.Open(url); err != nil {
		fmt.Println("[RTSP] Error", err)
	} else {
		for {
			select {
			case <-Client.Signals:
				fmt.Println("Exit signals by rtsp")
				return
			case data := <-Client.Outgoing:
				// fmt.Println("recive  rtp packet size", len(data))
				packet := &rtp.Packet{}
				err = packet.Unmarshal(data[4:])
				if err != nil {
					continue
				}
				if packet.PayloadType == 96 { //RTSP H264 PayloadType = 96
					if videoTrack != nil {
						packet.PayloadType = webrtc.DefaultPayloadTypeH264
						packet.SSRC = videoTrack.SSRC()
						videoTrack.WriteRTP(packet)
					}
				}
			}
		}
	}
	Client.Close()
}
