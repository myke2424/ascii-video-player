package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"sync"
	"time"

	"github.com/qeesung/image2ascii/convert"
)

// Frame represents a video single frame
type Frame struct {
	Image image.Image
}

// Struct to encapsulate anything related to frame rendering
type Renderer struct {
	width        int
	height       int
	frameRate    float64
	useGreyScale bool
	frameChannel chan Frame
}

// BufferFrames reads raw RGB data from stdout and sends each frame through the frame channel for rendering
func (r *Renderer) BufferFrames(stdout *bufio.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	frameSize := r.width * r.height * 3 // RGB format, each pixel is 3 bytes (1 byte per color)
	frameBuffer := make([]byte, frameSize)

	for {
		_, err := io.ReadFull(stdout, frameBuffer)
		if err != nil {
			if err == io.EOF {
				close(r.frameChannel)
				break
			}
			// If we fail to read a frame, continue to the next frame
			continue
		}

		frame := r.rawRGBToFrame(frameBuffer, r.width, r.height)
		r.frameChannel <- frame
	}
}

// RenderFrames reads buffered frames from the frame channel, converts them to ASCII and renders it to the terminal
func (r *Renderer) RenderFrames(start chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	converter := convert.NewImageConverter()
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = r.width
	convertOptions.FixedHeight = r.height
	convertOptions.Colored = !r.useGreyScale

	startTime := time.Now()
	frameDuration := time.Second / time.Duration(r.frameRate)
	<-start // Wait for the signal to start rendering - used to sync with audio

	frameIndex := 0
	for frame := range r.frameChannel {
		r.renderFrame(frame, converter, &convertOptions)
		// Calculate time to sleep until the next frame
		nextFrameTime := startTime.Add(frameDuration * time.Duration(frameIndex+1))
		time.Sleep(time.Until(nextFrameTime))
		frameIndex++
	}
}

// RenderFrame renders a single video frame as ASCII
func (r *Renderer) renderFrame(frame Frame, imageConverter *convert.ImageConverter, converterOptions *convert.Options) {
	asciiArt := imageConverter.Image2ASCIIString(frame.Image, converterOptions)
	fmt.Print("\033[H\033[2J") // Clear terminal escape sequence
	fmt.Println(asciiArt)
}

// rawRGBToImage converts raw RGB frame bytes to a Frame struct
func (r *Renderer) rawRGBToFrame(frame []byte, width, height int) Frame {
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
	return Frame{Image: img}
}
