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

func StartDownload(dm *DownloadManager, program *tea.Program, req types.DownloadRequest) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		unfinished := UnfinishedDownload{
			URL:       req.URL,
			FormatID:  req.FormatID,
			Title:     req.Title,
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

		cb := req.CookiesFromBrowser
		c := req.Cookies
		if cb == "" {
			cb = cfg.CookiesBrowser
		}
		if c == "" {
			c = cfg.CookiesFile
		}

		go doDownload(dm, program, req, downloadPath, cfg.YTDLPPath, cfg.FFmpegPath, cb, c)
		return nil
	})
}

func doDownload(dm *DownloadManager, program *tea.Program, req types.DownloadRequest, outputPath, ytDlpPath, ffmpegPath, cookiesBrowser, cookiesFile string) {
	ctx, cancel := context.WithCancel(context.Background())
	dm.SetContext(ctx, cancel)

	if ytDlpPath == "" {
		ytDlpPath = "yt-dlp"
	}
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	url := req.URL
	formatID := req.FormatID
	abr := req.ABR

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

	args = []string{
		"-f",
		formatID,
		"--newline",
		"-R",
		"infinite",
		url,
	}

	if req.IsAudioTab {
		audioQuality := fmt.Sprintf("%dK", int(abr))
		fileExtension = ".mp3"
		ext := strings.Replace(fileExtension, ".", "", -1)
		args = append([]string{
			"-o",
			filepath.Join(outputPath, "%(artist)s - %(title)s.%(ext)s"),
			"--restrict-filenames",
			"-x",
			"--audio-format",
			ext,
			"--audio-quality",
			audioQuality,
			"--add-metadata",
			"--metadata-from-title",
			"%(artist)s - %(title)s",
		}, args...)
	} else {
		fileExtension = ".mp4"
		ext := strings.Replace(fileExtension, ".", "", -1)
		args = append([]string{
			"-o",
			filepath.Join(outputPath, "%(title)s.%(ext)s"),
			"--merge-output-format",
			ext,
			"--remux-video",
			ext,
		}, args...)
	}

	if !isPlaylist {
		args = append([]string{"--no-playlist"}, args...)
	}

	if cookiesBrowser != "" {
		args = append([]string{"--cookies-from-browser", cookiesBrowser}, args...)
	} else if cookiesFile != "" {
		args = append([]string{"--cookies", cookiesFile}, args...)
	}

	for _, opt := range req.Options {
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

	log.Print("args: ", args)
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
		stdout.Close()
		log.Printf("stderr pipe error: %v", err2)
		errMsg := fmt.Sprintf("stderr pipe error: %v", err2)
		program.Send(types.DownloadResultMsg{Err: errMsg})
		return
	}

	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		log.Printf("start error: %v", err)
		errMsg := fmt.Sprintf("start error: %v", err)
		program.Send(types.DownloadResultMsg{Err: errMsg})
		return
	}

	parser := NewProgressParser()
	var wg sync.WaitGroup
	readPipe := func(pipe io.Reader) {
		defer wg.Done()
		parser.ReadPipe(pipe, func(percent float64, speed, eta, status, destination string) {
			program.Send(types.ProgressMsg{Percent: percent, Speed: speed, Eta: eta, Status: status, Destination: destination, FileExtension: fileExtension})
		})
	}

	wg.Add(2)
	go readPipe(stdout)
	go readPipe(stderr)
	wg.Wait()
	err = cmd.Wait()

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
