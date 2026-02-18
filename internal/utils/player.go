package utils

import (
	"log"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdagiz/xytz/internal/types"
)

var (
	currentMPV             *exec.Cmd
	mpvKilledIntentionally bool
)

func PlayURLWithMPV(url string, ytdlFormat string, video types.VideoItem, program *tea.Program) tea.Cmd {
	return func() tea.Msg {
		args := make([]string, 0, 2)
		if ytdlFormat != "" {
			args = append(args, "--ytdl-format="+ytdlFormat)
		}

		args = append(args, url)
		cmd := exec.Command("mpv", args...)

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to play video with mpv: %v", err)
			return types.PlayVideoMsg{ErrMsg: "Failed to play video with mpv"}
		}

		currentMPV = cmd
		go func() {
			err := cmd.Wait()

			if !mpvKilledIntentionally {
				if err != nil {
					log.Printf("mpv exited with error: %v", err)
				}

				currentMPV = nil
				if program != nil {
					program.Send(types.PlayVideoMsg{SelectedVideo: video})
				}
			} else {
				currentMPV = nil
				mpvKilledIntentionally = false
			}
		}()

		return types.MPVStartedMsg{SelectedVideo: video}
	}
}

func KillMPV() {
	if currentMPV != nil {
		mpvKilledIntentionally = true
		if err := currentMPV.Process.Kill(); err != nil {
			mpvKilledIntentionally = false
		}

		currentMPV = nil
	}
}

func IsMPVRunning() bool {
	if currentMPV == nil {
		return false
	}

	err := currentMPV.Process.Signal(syscall.Signal(0))
	return err == nil
}
