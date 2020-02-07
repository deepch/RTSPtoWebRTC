package main

import (
	"log"
	"time"

	"github.com/deepch/vdk/format/rtsp"
)

func serveStreams() {
	for k, v := range Config.Streams {
		go func(name, url string) {
			for {
				log.Println(name, "connect", url)
				rtsp.DebugRtsp = true
				session, err := rtsp.Dial(url)
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				session.RtpKeepAliveTimeout = time.Duration(10 * time.Second)
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				codec, err := session.Streams()
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				Config.coAd(name, codec)
				for {
					pkt, err := session.ReadPacket()
					if err != nil {
						log.Println(name, err)
						break
					}
					Config.cast(name, pkt)
				}
				session.Close()
				log.Println(name, "reconnect wait 5s")
				time.Sleep(5 * time.Second)
			}
		}(k, v.URL)
	}
}
