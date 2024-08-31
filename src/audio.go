package main

import (
	"fmt"
	"os/exec"
	"sync"
)

// PlayAudio plays the audio using ffplay
func PlayAudio(videoFilePath string, start chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Wait for the start signal to start playing audio
	<-start

	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-i", videoFilePath)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error playing audio:", err)
	}
}
