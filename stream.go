package main

import (
	"errors"
	"log"
	"time"

	"github.com/deepch/vdk/format/rtspv2"
)

var (
	ErrorStreamExitNoVideoOnStream = errors.New("Stream Exit No Video On Stream")
	ErrorStreamExitRtspDisconnect  = errors.New("Stream Exit Rtsp Disconnect")
)

func serveStreams() {
	for k, v := range Config.Streams {
		go RTSPWorkerLoop(k, v.URL)
	}
}
func RTSPWorkerLoop(name, url string) {
	for {
		err := RTSPWorker(name, url)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func RTSPWorker(name, url string) error {
	keyTest := time.NewTimer(20 * time.Second)
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: url, DisableAudio: true, DialTimeout: 3 * time.Second, ReadWriteTimeout: 3 * time.Second, Debug: false})
	if err != nil {
		return err
	}
	defer RTSPClient.Close()
	if RTSPClient.CodecData != nil {
		Config.coAd(name, RTSPClient.CodecData)
	}
	for {
		select {
		case <-keyTest.C:
			return ErrorStreamExitNoVideoOnStream
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				Config.coAd(name, RTSPClient.CodecData)
			case rtspv2.SignalStreamRTPStop:
				return ErrorStreamExitRtspDisconnect
			}
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
			}
			Config.cast(name, *packetAV)
		}
	}
}
