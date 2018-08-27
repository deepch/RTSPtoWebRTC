package main

import (
	"fmt"
	"log"
	"net/http"

	"encoding/base64"

	"github.com/gorilla/mux"
	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/ice"
)

var DataChanelTest chan<- webrtc.RTCSample

func StartHTTPServer() {
	r := mux.NewRouter()
	r.HandleFunc("/recive", HTTPHome)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	go func() {
		err := http.ListenAndServe(":8080", r)
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
	webrtc.RegisterDefaultCodecs()
	peerConnection, err := webrtc.New(webrtc.RTCConfiguration{
		ICEServers: []webrtc.RTCICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	peerConnection.OnICEConnectionStateChange = func(connectionState ice.ConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	}
	vp8Track, err := peerConnection.NewRTCTrack(webrtc.DefaultPayloadTypeH264, "video", "pion2")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = peerConnection.AddTrack(vp8Track)
	if err != nil {
		log.Println(err)
		return
	}
	offer := webrtc.RTCSessionDescription{
		Type: webrtc.RTCSdpTypeOffer,
		Sdp:  string(sd),
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
	w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(answer.Sdp))))
	DataChanelTest = vp8Track.Samples
}
