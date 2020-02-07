package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"

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
		//sps, pps := []byte{}, []byte{}
		sps := codecs[0].(h264parser.CodecData).SPS()
		pps := codecs[0].(h264parser.CodecData).PPS()
		//log.Fatalln(ch, codecs)
		sd, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(sd))
		// webrtc.RegisterDefaultCodecs()
		//peerConnection, err := webrtc.New(webrtc.RTCConfiguration{
		//var m webrtc.MediaEngine
		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		})
		if err != nil {
			panic(err)
		}
		videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", suuid+"_video")
		if err != nil {
			log.Println(err)
			return
		}
		_, err = peerConnection.AddTrack(videoTrack)
		if err != nil {
			log.Println(err)
			return
		}
		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  string(sd),
		}
		if err := peerConnection.SetRemoteDescription(offer); err != nil {
			log.Println(err)
			return
		}
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			log.Println(err)
			return
		}
		c.Writer.Write([]byte(base64.StdEncoding.EncodeToString([]byte(answer.SDP))))
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
			var pre uint32
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
					var ts uint32
					if pre != 0 {
						ts = uint32(timeToTs(pck.Time)) - pre
					}
					err := videoTrack.WriteSample(media.Sample{Data: pck.Data, Samples: uint32(ts)})
					pre = uint32(timeToTs(pck.Time))
					if err != nil {
						return
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
