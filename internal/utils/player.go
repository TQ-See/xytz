package utils

import (
	"log"
	"os/exec"
)

var PlayURLWithMPVFunc = playURLWithMPV

func PlayURLWithMPV(url string, ytdlFormat string) {
	PlayURLWithMPVFunc(url, ytdlFormat)
}

func playURLWithMPV(url string, ytdlFormat string) {
	go func() {
		args := make([]string, 0, 2)
		if ytdlFormat != "" {
			args = append(args, "--ytdl-format="+ytdlFormat)
		}

		args = append(args, url)
		cmd := exec.Command("mpv", args...)

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to play video with mpv: %v", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			log.Printf("mpv exited with error: %v", err)
		}
	}()
}
