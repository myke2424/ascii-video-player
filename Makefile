all: install build

install-ffmpeg:
	sudo apt-get update
	sudo apt-get install -y ffmpeg

install-mediainfo:
	sudo apt-get update
	sudo apt-get install -y mediainfo

install: install-ffmpeg install-mediainfo

install-dev: install
	go mod tidy
	go mod download

build: 
	go build -o ascii-player src/main.go src/video.go src/audio.go

clean:
	rm -f ascii-player

# Declare phony targets to avoid conflicts with files
.PHONY: all install install-ffmpeg install-mediainfo install-dev build clean

