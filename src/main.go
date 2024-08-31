package main

import (
	"bufio"
	"bytes"
	"flag"
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

// CLI Args
type Config struct {
	Video string
	Grey  bool
}

// Get the terminal size dimensions. If we fail to obtain them, use the defaults provided.
func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 160, 90
	}
	return width, height
}

// Get the video frame rate using MediaInfo. If we fail to find the framerate, use a default of 30.
func getVideoFrameRate(videoFilePath string) float64 {
	cmd := exec.Command("mediainfo", "--Inform=Video;%FrameRate%", videoFilePath)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 30
	}

	frameRateStr := strings.TrimSpace(stdout.String())
	frameRate, err := strconv.ParseFloat(frameRateStr, 64)
	if err != nil {
		return 30
	}

	return frameRate
}

// Executes an FFmpeg command to read a video file, downscales it to specified dimensions to fit the terminal,
// convert it to raw RGB format, and output the raw video data to stdout.
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

// Convert raw RGB frame data to an image.Image object
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

// Read the raw RGB pixel data from stdout and convert it to ASCII frames using image2ascii
func RawRGBToASCII(stdout *bufio.Reader, frameRate float64, width, height int, grey bool, start chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	frameSize := width * height * 3 // RGB format, each pixel is 3 bytes (1 byte per color)
	frameBuffer := make([]byte, frameSize)

	converter := convert.NewImageConverter()
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = width
	convertOptions.FixedHeight = height
	convertOptions.Colored = !grey

	// Wait for the start signal to start processing video frames
	<-start

	startTime := time.Now()
	frameDuration := time.Second / time.Duration(frameRate)

	frameIndex := 0
	for {
		_, err := io.ReadFull(stdout, frameBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			// If we fail to read a frame, continue to the next frame
			continue
		}

		img := rawRGBToImage(frameBuffer, width, height)
		asciiArt := converter.Image2ASCIIString(img, &convertOptions)

		fmt.Print("\033[H\033[2J") // Clear terminal escape sequence
		fmt.Println(asciiArt)

		// Calculate time to sleep until the next frame
		nextFrameTime := startTime.Add(frameDuration * time.Duration(frameIndex+1))
		time.Sleep(time.Until(nextFrameTime))

		frameIndex++
	}
}

// Plays the audio using ffplay. It waits for the start signal before beginning playback.
func playAudio(videoFilePath string, start chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Wait for the start signal to start playing audio
	<-start

	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-i", videoFilePath)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error playing audio:", err)
	}
}

func main() {
	var config Config
	flag.StringVar(&config.Video, "video", "", "Video file path you want to playout - required")
	flag.BoolVar(&config.Grey, "grey", false, "Render greyscale. If not passed in, use RGB  - not required.")
	flag.Parse()

	if len(config.Video) == 0 {
		panic("--video flag is required. Please provide a file path to the video you want to playout using this flag")
	}

	width, height := getTerminalSize()
	cmd, stdout, err := VideoToRawRGB(config.Video, width, height)

	if err != nil {
		panic(err)
	}

	frameRate := getVideoFrameRate(config.Video)
	var wg sync.WaitGroup
	wg.Add(2)

	start := make(chan struct{})
	go RawRGBToASCII(stdout, frameRate, width, height, config.Grey, start, &wg)
	go playAudio(config.Video, start, &wg)

	// Close the start channel to signal both video/audio goroutines to start.
	// This will ensure video frame processing / audio playback start at the same time and are in sync.
	close(start)

	wg.Wait()
	cmd.Wait()
}
