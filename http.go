package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"encoding/base64"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v2"
	//ice "github.com/pions/webrtc/internal/ice"
)

// var DataChanelTest chan<- webrtc.RTCSample
var videoTrack *webrtc.Track

func StartHTTPServer() {
	r := mux.NewRouter()
	r.HandleFunc("/recive", HTTPHome)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	go func() {
		err := http.ListenAndServe(":25000", r)
		if err != nil {
		}
	}()
	select {}
}
func HTTPHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := r.FormValue("data")
	sd, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Println(err)
		return
	}
	// webrtc.RegisterDefaultCodecs()
	//peerConnection, err := webrtc.New(webrtc.RTCConfiguration{
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
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})
	vp8Track, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion2")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = peerConnection.AddTrack(vp8Track)
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
	w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(answer.SDP))))
	// DataChanelTest = vp8Track.Samples
	videoTrack = vp8Track
}
