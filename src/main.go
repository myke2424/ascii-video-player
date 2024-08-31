package main

import (
	"flag"
	"log"
	"sync"
)

type Config struct {
	Video string
	Grey  bool
}

func (c *Config) ParseCliArgs() {
	flag.StringVar(&c.Video, "video", "", "Video file path you want to playout - required")
	flag.BoolVar(&c.Grey, "grey", false, "Render greyscale. If not passed in, use RGB - not required.")
	flag.Parse()

	if len(c.Video) == 0 {
		log.Fatal("--video flag is required. Please provide a file path to the video you want to playout using this flag")
	}
}

func main() {
	var cfg Config
	cfg.ParseCliArgs()
	width, height := GetTerminalSize()

	videoCmd, stdout, err := VideoToRawRGB(cfg.Video, width, height)
	if err != nil {
		log.Fatalf("Failed to initialize video processing: %v", err)
	}
	defer videoCmd.Wait()
	frameRate := GetVideoFrameRate(cfg.Video)

	var wg sync.WaitGroup
	start := make(chan struct{}) // channel to enforce synchronization between video/audio
	frameBuffer := make(chan Frame, 10)

	wg.Add(3)

	go BufferFrames(stdout, width, height, frameBuffer, &wg)
	go RenderFrames(frameBuffer, frameRate, width, height, cfg.Grey, start, &wg)
	go PlayAudio(cfg.Video, start, &wg)

	// Close the start channel to signal both video/audio goroutines to start.
	// This will ensure rendering frames and audio playback are in sync
	close(start)
	wg.Wait()
}
