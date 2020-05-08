package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

func serveHTTP() {
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/", func(c *gin.Context) {
		fi, all := Config.list()
		sort.Strings(all)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"port":     Config.Server.HTTPPort,
			"suuid":    fi,
			"suuidMap": all,
			"version":  time.Now().String(),
		})
	})
	router.GET("/player/:suuid", func(c *gin.Context) {
		_, all := Config.list()
		sort.Strings(all)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"port":     Config.Server.HTTPPort,
			"suuid":    c.Param("suuid"),
			"suuidMap": all,
			"version":  time.Now().String(),
		})
	})
	router.POST("/recive", reciver)
	router.StaticFS("/static", http.Dir("web/static"))
	err := router.Run(Config.Server.HTTPPort)
	if err != nil {
		log.Fatalln(err)
	}
}
func reciver(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	data := c.PostForm("data")
	suuid := c.PostForm("suuid")
	log.Println("Request", suuid)
	if Config.ext(suuid) {
		codecs := Config.coGe(suuid)
		if codecs == nil {
			log.Println("No Codec Info")
			return
		}
		sps := codecs[0].(h264parser.CodecData).SPS()
		pps := codecs[0].(h264parser.CodecData).PPS()
		sd, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			log.Println("decode SDP error", err)
			return
		}

		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
			//	ICEServers: []webrtc.ICEServer{
			//		{
			//URLs: []string{"stun:stun.l.google.com:19302"},
			//		},
			///	},
		})
		if err != nil {
			panic(err)
		}
		videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", suuid+"video")
		_, err = peerConnection.AddTrack(videoTrack)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = peerConnection.AddTransceiverFromTrack(videoTrack,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			},
		)
		payloadType := videoTrack.PayloadType
		log.Fatalln(payloadType)
		var audioTrack *webrtc.Track
		if len(codecs) > 1 && (codecs[1].Type() == av.PCM_ALAW || codecs[1].Type() == av.PCM_MULAW) {
			switch codecs[1].Type() {
			case av.PCM_ALAW:
				audioTrack, err = peerConnection.NewTrack(webrtc.DefaultPayloadTypePCMA, rand.Uint32(), "audio", suuid+"audio")
			case av.PCM_MULAW:
				audioTrack, err = peerConnection.NewTrack(webrtc.DefaultPayloadTypePCMU, rand.Uint32(), "audio", suuid+"audio")
			}
			if err != nil {
				log.Println(err)
				return
			}
			_, err = peerConnection.AddTrack(audioTrack)
			if err != nil {
				log.Println(err)
				return
			}
		}
		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  string(sd),
		}
		if err := peerConnection.SetRemoteDescription(offer); err != nil {
			log.Println("SetRemoteDescription error", err, string(offer.SDP))
			return
		}
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			log.Println("CreateAnswer error", err)
			return
		}
		//spd_string := string(answer.SDP)
		//strings.
		deep := strings.Replace(string(answer.SDP), "192.168.33.33", "171.25.233.50", -1)
		//deep = strings.Replace(string(deep), "192.168.33.33", "Sendonly", -1)
		//Sendonly
		//log.Println(string(answer.SDP), string(deep))
		c.Writer.Write([]byte(base64.StdEncoding.EncodeToString([]byte(deep))))
		go func() {
			control := make(chan bool, 10)
			conected := make(chan bool, 10)
			defer peerConnection.Close()
			peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
				fmt.Printf("Connection State has changed %s \n", connectionState.String())
				if connectionState != webrtc.ICEConnectionStateConnected {
					log.Println("Client Close Exit")
					control <- true
					return
				}
				if connectionState == webrtc.ICEConnectionStateConnected {
					conected <- true
				}
			})
			<-conected
			cuuid, ch := Config.clAd(suuid)
			defer Config.clDe(suuid, cuuid)
			var Vpre time.Duration

			var start bool
			for {
				select {
				case <-control:
					return
				case pck := <-ch:
					if pck.IsKeyFrame {
						start = true
					}
					if !start {
						continue
					}
					if pck.IsKeyFrame {
						pck.Data = append([]byte("\000\000\001"+string(sps)+"\000\000\001"+string(pps)+"\000\000\001"), pck.Data[4:]...)

					} else {
						pck.Data = pck.Data[4:]
					}
					var Vts time.Duration
					if pck.Idx == 0 && videoTrack != nil {
						if Vpre != 0 {
							Vts = pck.Time - Vpre
						}
						samples := uint32(90000 / 1000 * Vts.Milliseconds())
						err := videoTrack.WriteSample(media.Sample{Data: pck.Data, Samples: uint32(samples)})
						if err != nil {
							return
						}
						Vpre = pck.Time
					} else if pck.Idx == 1 && audioTrack != nil {
						err := audioTrack.WriteSample(media.Sample{Data: pck.Data, Samples: uint32(len(pck.Data))})
						if err != nil {
							return
						}
					}
				}
			}
		}()
		return
	}
}
func timeToTs(tm time.Duration) int64 {
	return int64(tm * time.Duration(90000) / time.Second)
}
