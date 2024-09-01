package main

import (
	"flag"
	"log"
	"os"
	"sync"
)

type Config struct {
	VideoPath    string
	UseGreyScale bool
}

func (c *Config) ParseCliArgs() {
	flag.StringVar(&c.VideoPath, "video", "", "Video file path you want to playout - required")
	flag.BoolVar(&c.UseGreyScale, "grey", false, "Render greyscale. If not passed in, use RGB - not required.")
	flag.Parse()

	if len(c.VideoPath) == 0 {
		log.Fatal("--video flag is required. Please provide a file path to the video you want to playout using this flag")
	}

	_, err := os.Stat(c.VideoPath)

	if os.IsNotExist(err) {
		log.Fatalf("The given video file path does not exist: [%s]", c.VideoPath)
	}
}

func main() {
	var cfg Config
	cfg.ParseCliArgs()
	width, height := GetTerminalSize()

	videoCmd, stdout, err := VideoToRawRGB(cfg.VideoPath, width, height)
	if err != nil {
		log.Fatalf("Failed to initialize video processing: %v", err)
	}
	defer videoCmd.Wait()
	frameRate := GetVideoFrameRate(cfg.VideoPath)

	var wg sync.WaitGroup
	start := make(chan struct{}) // channel to enforce synchronization between video/audio
	frameBuffer := make(chan Frame, 10)

	wg.Add(3)

	go BufferFrames(stdout, width, height, frameBuffer, &wg)
	go RenderFrames(frameBuffer, frameRate, width, height, cfg.UseGreyScale, start, &wg)
	go PlayAudio(cfg.VideoPath, start, &wg)

	// Close the start channel to signal both video/audio goroutines to start.
	// This will ensure rendering frames and audio playback are in sync
	close(start)
	wg.Wait()
}
