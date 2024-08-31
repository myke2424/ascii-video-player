package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qeesung/image2ascii/convert"
	"golang.org/x/term"
)

// Frame represents a video frame
type Frame struct {
	Image image.Image
}

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

// VideoToRawRGB executes an FFmpeg command to read video data and convert it to raw RGB format
func VideoToRawRGB(videoFilePath string, width, height int) (*exec.Cmd, *bufio.Reader, error) {
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

// BufferFrames reads raw RGB data from stdout and stores it in a buffer
func BufferFrames(stdout *bufio.Reader, width, height int, buffer chan Frame, wg *sync.WaitGroup) {
	defer wg.Done()

	frameSize := width * height * 3 // RGB format, each pixel is 3 bytes (1 byte per color)
	frameBuffer := make([]byte, frameSize)

	for {
		_, err := io.ReadFull(stdout, frameBuffer)
		if err != nil {
			if err == io.EOF {
				close(buffer)
				break
			}
			// If we fail to read a frame, continue to the next frame
			continue
		}

		img := rawRGBToImage(frameBuffer, width, height)
		buffer <- Frame{Image: img}
	}
}

// rawRGBToImage converts raw RGB frame data to an image.Image object
func rawRGBToImage(frame []byte, width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 3
			r := frame[offset]
			g := frame[offset+1]
			b := frame[offset+2]
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// RenderFrames reads buffered frames and converts them to ASCII
func RenderFrames(buffer chan Frame, frameRate float64, width, height int, grey bool, start chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	converter := convert.NewImageConverter()
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = width
	convertOptions.FixedHeight = height
	convertOptions.Colored = !grey

	startTime := time.Now()
	frameDuration := time.Second / time.Duration(frameRate)
	<-start // Wait for the signal to start rendering - used to sync with audio

	frameIndex := 0
	for frame := range buffer {
		asciiArt := converter.Image2ASCIIString(frame.Image, &convertOptions)

		fmt.Print("\033[H\033[2J") // Clear terminal escape sequence
		fmt.Println(asciiArt)

		// Calculate time to sleep until the next frame
		nextFrameTime := startTime.Add(frameDuration * time.Duration(frameIndex+1))
		time.Sleep(time.Until(nextFrameTime))

		frameIndex++
	}
}
