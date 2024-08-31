all: install

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

.PHONY: all install install-ffmpeg install-mediainfo install-dev

