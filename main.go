package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// TODO: Extract resolution from the video? and Frame rate?

// Get the terminal size dimensions. If we failed to obtain them, use the defaults provided.
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

// Executes an FFmpeg command to read a video file, resize it to specified dimensions to fit the terminal,
// convert it to raw RGB format, and output the raw video data to stdout.
func convertVideoToRawRGB(videoFilePath string, width int, height int) (*exec.Cmd, io.ReadCloser, error) {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-f", "rawvideo", "-pix_fmt", "rgb24", "-vf", fmt.Sprintf("scale=%d:%d", width, height), "-")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, nil, fmt.Errorf("error getting the stdout pipe for the ffmpeg command: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("error starting ffmpeg command to decode video to raw pixel data: %v", err)
	}

	return cmd, stdout, nil

}

// TODO: 8x8 characters?, pixel 1x1, text8x8
// downscale video, divide image by resolution of text (divide by 8)
// quantize luminance to smaller set of values (10 values maybe)?
// edge detective (sobel filter, canny edge, difference of guassaisns)?
// Compute shader?
// edge-detector???
// use lower contrest color combo instead of black and white
// depth of field effect

// Converts a raw RGB frame to an ASCII art string
func convertFrameToASCII(frame []byte, width int, height int) string {
	// More detailed ASCII characters from dark to light
	asciiChars := " .`-^\",:;Il!i~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

	var buffer bytes.Buffer
	for i := 0; i < len(frame); i += 3 {
		r := frame[i]
		g := frame[i+1]
		b := frame[i+2]

		// Calculate the grayscale value with gamma correction
		gray := 0.299*math.Pow(float64(r)/255.0, 2.2) +
			0.587*math.Pow(float64(g)/255.0, 2.2) +
			0.114*math.Pow(float64(b)/255.0, 2.2)
		gray = math.Pow(gray, 1.0/2.2)

		// Map grayscale to an ASCII character
		asciiIndex := int(gray * float64(len(asciiChars)-1))
		buffer.WriteByte(asciiChars[asciiIndex])

		// Add newline at the end of each row
		if (i/3+1)%width == 0 {
			buffer.WriteByte('\n')
		}
	}
	return buffer.String()
}

// Read the raw pixel data from stdout and convert to ASCII frames
func convertRawPixelDataToASCII(stdout io.ReadCloser, frameRate float64, width int, height int) {
	reader := bufio.NewReader(stdout)
	frameSize := width * height * 3 // RGB format, each pixel is 3 bytes (1 byte per color)
	frameBuffer := make([]byte, frameSize)

	for {
		_, err := io.ReadFull(reader, frameBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			// if we fail to read a frame, continue to the next frame
			continue
		}

		asciiFrame := convertFrameToASCII(frameBuffer, width, height)

		fmt.Print("\033[H\033[2J") // Clear terminal escape sequence
		fmt.Println(asciiFrame)

		// Wait for the next frame
		time.Sleep(time.Second / time.Duration(frameRate))
	}
}

func main() {
	var videoFilePath string
	flag.StringVar(&videoFilePath, "video", "", "Video file path you want to playout")
	flag.Parse()

	if len(videoFilePath) == 0 {
		panic("--video flag is required. Please provide a file path to the video you want to playout using this flag")
	}

	width, height := getTerminalSize()
	cmd, stdout, err := convertVideoToRawRGB(videoFilePath, width, height)

	if err != nil {
		panic(err)
	}

	frameRate := getVideoFrameRate(videoFilePath)
	convertRawPixelDataToASCII(stdout, frameRate, width, height)
	cmd.Wait()
}
