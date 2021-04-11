@REM https://habr.com/ru/post/249449/

@SET GOOS=windows
@SET GOARCH=amd64
go build -ldflags "-s -w" -o rtsp2webrtc_amd64.exe

@SET GOOS=linux
@SET GOARCH=amd64
go build -ldflags "-s -w" -o rtsp2webrtc_amd64

@SET GOOS=linux
@SET GOARCH=arm
@SET GOARM=7
go build -ldflags "-s -w" -o rtsp2webrtc_armv7

@SET GOOS=linux
@SET GOARCH=arm64
go build -ldflags "-s -w" -o rtsp2webrtc_aarch64
