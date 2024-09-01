package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// GetTerminalSize returns the terminal dimensions or defaults
func GetTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return width, height
}

// GetVideoFrameRate returns the frame rate of a video file
func GetVideoFrameRate(videoFilePath string) float64 {
	cmd := exec.Command("mediainfo", "--Inform=Video;%FrameRate%", videoFilePath)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 30.0
	}

	frameRateStr := strings.TrimSpace(stdout.String())
	frameRate, err := strconv.ParseFloat(frameRateStr, 64)
	if err != nil {
		return 30.0
	}

	return frameRate
}

// DecodeVideoToRawRGB executes an FFmpeg command to read video data and convert it to raw RGB format
func DecodeVideoToRawRGB(videoFilePath string, width, height int) (*exec.Cmd, *bufio.Reader, error) {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-f", "rawvideo", "-pix_fmt", "rgb24", "-vf", fmt.Sprintf("scale=%d:%d", width, height), "-")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, nil, fmt.Errorf("error getting the stdout pipe for the ffmpeg command: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("error starting ffmpeg command to decode video to raw pixel data: %v", err)
	}

	return cmd, bufio.NewReader(stdout), nil
}
