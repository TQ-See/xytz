//go:build windows

package utils

import (
	"log"

	"github.com/xdagiz/xytz/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func PauseDownload(dm *DownloadManager) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		cmd := dm.GetCmd()
		if cmd != nil && cmd.Process != nil && !dm.IsPaused() {
			log.Print("pause not supported on windows")
		}

		return types.PauseDownloadMsg{}
	})
}

func ResumeDownload(dm *DownloadManager) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		cmd := dm.GetCmd()
		if cmd != nil && cmd.Process != nil && dm.IsPaused() {
			log.Print("resume not supported on windows")
		}

		return types.ResumeDownloadMsg{}
	})
}
