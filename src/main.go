package main

import (
	"flag"
	"log"
	"os"
	"sync"
)

// Config contains all the CLI args
type Config struct {
	VideoPath    string
	UseGreyScale bool
}

// Parse the cli args, validate the file exists that we want to process
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

/*
Runs the ASCII video player.

It decodes the video to raw RGB, buffers the frames, renders them as ASCII art,
and plays the audio in sync with the video. Uses a start channel to kick off
video rendering and audio playback at the same time.
*/
func main() {
	var cfg Config
	cfg.ParseCliArgs()
	width, height := GetTerminalSize()

	videoCmd, stdout, err := DecodeVideoToRawRGB(cfg.VideoPath, width, height)
	if err != nil {
		log.Fatalf("Failed to initialize video processing: %v", err)
	}
	defer videoCmd.Wait()
	frameRate := GetVideoFrameRate(cfg.VideoPath)

	var wg sync.WaitGroup
	wg.Add(3)

	start := make(chan struct{})         // channel to enforce synchronization between video/audio
	frameChannel := make(chan Frame, 10) // channel used to buffer frames for rendering
	renderer := Renderer{width: width, height: height, frameRate: frameRate, useGreyScale: cfg.UseGreyScale, frameChannel: frameChannel}

	go renderer.BufferFrames(stdout, &wg)
	go renderer.RenderFrames(start, &wg)
	go PlayAudio(cfg.VideoPath, start, &wg)

	// Close the start channel to signal both video/audio goroutines to start.
	// This will ensure rendering frames and audio playback are in sync
	close(start)
	wg.Wait()
}
