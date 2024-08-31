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
	// Parse CLI arguments
	var cfg Config
	cfg.ParseCliArgs()

	width, height := GetTerminalSize()

	// Initialize video processing
	videoCmd, stdout, err := VideoToRawRGB(cfg.Video, width, height)
	if err != nil {
		log.Fatalf("Failed to initialize video processing: %v", err)
	}
	defer videoCmd.Wait()

	frameRate := GetVideoFrameRate(cfg.Video)

	// Set up synchronization
	var wg sync.WaitGroup
	start := make(chan struct{})

	// Frame buffer channel
	buffer := make(chan Frame, 10)

	wg.Add(3)

	// Start video, ASCII, and audio processing in separate goroutines
	go BufferFrames(stdout, width, height, buffer, &wg)
	go RenderFrames(buffer, frameRate, width, height, cfg.Grey, &wg)
	go PlayAudio(cfg.Video, start, &wg)

	close(start)

	// Wait for all goroutines to finish
	wg.Wait()
}
