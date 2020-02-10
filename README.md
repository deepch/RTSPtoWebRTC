# RTSPtoWebRTC

RTSP Stream to WebBrowser over WebRTC based on Pion

full native! not use ffmpeg or gstreamer

if you need RTSPtoWSMP4f use https://github.com/deepch/RTSPtoWSMP4f

## Team

Deepch - https://github.com/deepch streaming developer

Dmitry - https://github.com/vdalex25 web developer

![RTSPtoWebRTC image](doc/demo4.png)

## Installation
1.
```bash
go get github.com/deepch/RTSPtoWebRTC
```
2.
```bash
cd src/github.com/deepch/RTSPtoWebRTC
```
3.
```bash
go run *.go
```
4.
```bash
open web browser http://127.0.0.1:8083 work google chrome
```

## Configuration

### Edit file config.json

format:

```bash
{
  "server": {
    "http_port": ":8083"
  },
  "streams": {
    "demo1": {
      "url": "rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa"
    },
    "demo2": {
      "url": "rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa"
    },
    "demo3": {
      "url": "rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa"
    }
  }
}
```

## Limitations

Video Codecs Supported: H264

Audio Codecs Supported: none
