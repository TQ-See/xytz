package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func StartDownload(dm *DownloadManager, program *tea.Program, url, formatID string, isAudioTab bool, abr float64, title string, options []types.DownloadOption, cookiesFromBrowser, cookies string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		unfinished := UnfinishedDownload{
			URL:       url,
			FormatID:  formatID,
			Title:     title,
			Timestamp: time.Now(),
		}

		if err := AddUnfinished(unfinished); err != nil {
			log.Printf("Failed to add to unfinished list: %v", err)
		}

		cfg, err := config.Load()
		if err != nil {
			log.Printf("Warning: Failed to load config, using defaults: %v", err)
			cfg = config.GetDefault()
		}

		downloadPath := cfg.GetDownloadPath()

		cb := cookiesFromBrowser
		c := cookies
		if cb == "" {
			cb = cfg.CookiesBrowser
		}
		if c == "" {
			c = cfg.CookiesFile
		}

		go doDownload(dm, program, url, formatID, isAudioTab, abr, downloadPath, cfg.YTDLPPath, cb, c, options)

		return nil
	})
}

func CancelDownload(dm *DownloadManager) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if err := dm.Cancel(); err != nil {
			log.Printf("Failed to cancel download: %v", err)
		}

		return types.CancelDownloadMsg{}
	})
}

func doDownload(dm *DownloadManager, program *tea.Program, url, formatID string, isAudioTab bool, abr float64, outputPath, ytDlpPath, cookiesBrowser, cookiesFile string, options []types.DownloadOption) {
	ctx, cancel := context.WithCancel(context.Background())
	dm.SetContext(ctx, cancel)

	if ytDlpPath == "" {
		ytDlpPath = "yt-dlp"
	}

	if url == "" {
		log.Printf("download error: empty URL provided")
		program.Send(types.DownloadResultMsg{Err: "Download error: empty URL provided"})
		return
	}

	isPlaylist := strings.Contains(url, "/playlist?list=") || strings.Contains(url, "&list=")

	var (
		args          []string
		fileExtension string
	)

	if isAudioTab {
		audioQuality := fmt.Sprintf("%dK", int(abr))
		fileExtension = ".mp3"
		args = []string{
			"-f",
			formatID,
			"-o",
			filepath.Join(outputPath, "%(artist)s - %(title)s.%(ext)s"),
			"--restrict-filenames",
			"--embed-thumbnail",
			"-x",
			"--audio-format",
			"mp3",
			"--audio-quality",
			audioQuality,
			"--add-metadata",
			"--metadata-from-title",
			"%(artist)s - %(title)s",
			"--newline",
			"-R",
			"infinite",
			url,
		}
	} else {
		fileExtension = ".mp4"
		args = []string{
			"-f",
			formatID,
			"--newline",
			"-R",
			"infinite",
			"-o",
			filepath.Join(outputPath, "%(title)s.%(ext)s"),
			url,
		}
	}

	if !isPlaylist {
		args = append([]string{"--no-playlist"}, args...)
	}

	if cookiesBrowser != "" {
		args = append([]string{"--cookies-from-browser", cookiesBrowser}, args...)
	} else if cookiesFile != "" {
		args = append([]string{"--cookies", cookiesFile}, args...)
	}

	for _, opt := range options {
		if opt.Enabled {
			switch opt.ConfigField {
			case "EmbedSubtitles":
				args = append(args, "--embed-subs")
			case "EmbedMetadata":
				args = append(args, "--embed-metadata")
			case "EmbedChapters":
				args = append(args, "--embed-chapters")
			}
		}
	}

	cmd := exec.CommandContext(ctx, ytDlpPath, args...)

	dm.SetCmd(cmd)
	dm.SetPaused(false)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("pipe error: %v", err)
		errMsg := fmt.Sprintf("pipe error: %v", err)
		program.Send(types.DownloadResultMsg{Err: errMsg})
		return
	}

	stderr, err2 := cmd.StderrPipe()
	if err2 != nil {
		log.Printf("stderr pipe error: %v", err2)
		errMsg := fmt.Sprintf("stderr pipe error: %v", err2)
		program.Send(types.DownloadResultMsg{Err: errMsg})
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("start error: %v", err)
		errMsg := fmt.Sprintf("start error: %v", err)
		program.Send(types.DownloadResultMsg{Err: errMsg})
		return
	}

	parser := NewProgressParser()
	var wg sync.WaitGroup
	readPipe := func(pipe io.Reader) {
		wg.Add(1)
		defer wg.Done()
		parser.ReadPipe(pipe, func(percent float64, speed, eta, status, destination string) {
			program.Send(types.ProgressMsg{Percent: percent, Speed: speed, Eta: eta, Status: status, Destination: destination, FileExtension: fileExtension})
		})
	}

	go readPipe(stdout)
	go readPipe(stderr)
	wg.Wait()
	err = cmd.Wait()

	if stdout != nil {
		if err := stdout.Close(); err != nil {
			log.Printf("failed to close progress stdout: %v", err)
		}
	}

	if stderr != nil {
		if err := stderr.Close(); err != nil {
			log.Printf("failed to close progress stderr; %v", err)
		}
	}

	dm.Clear()

	if ctx.Err() == context.Canceled {
		program.Send(types.DownloadResultMsg{Err: "Download cancelled"})
		return
	}

	if err != nil {
		errMsg := fmt.Sprintf("Download error: %v", err)
		program.Send(types.DownloadResultMsg{Err: errMsg})
	} else {
		if err := RemoveUnfinished(url); err != nil {
			log.Printf("Failed to remove from unfinished list: %v", err)
		}

		program.Send(types.DownloadResultMsg{Output: "Download complete"})
	}
}
