package utils

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/types"
)

func runYTDLPCommand(sm *SearchManager, ytDlpPath, searchURL string, searchLimit int, args []string) ([]list.Item, []string, int, string, bool) {
	playlistItems := fmt.Sprintf("1:%d", searchLimit)
	cmdArgs := append(append([]string{}, args...),
		"--flat-playlist",
		"--dump-json",
		"--playlist-items", playlistItems,
		searchURL,
	)

	cmd := exec.Command(ytDlpPath, cmdArgs...)

	sm.SetCmd(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get stdout pipe: %v", err)
		return nil, nil, 0, errMsg, false
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get stderr pipe: %v", err)
		return nil, nil, 0, errMsg, false
	}

	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("failed to start search: %v", err)
		return nil, nil, 0, errMsg, false
	}

	var videos []list.Item
	skippedLiveShort := 0

	scanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	stderrLines := []string{}
	var stderrWg sync.WaitGroup
	stderrWg.Go(func() {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			stderrLines = append(stderrLines, line)
			log.Printf("yt-dlp stderr: %s", line)
		}
	})

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		videoItem, err := ParseVideoItem(trimmedLine)
		if err != nil {
			if errors.Is(err, ErrSkippedLiveShort) {
				skippedLiveShort++
				continue
			}

			log.Printf("Failed to parse video item: %v", err)
			continue
		}

		videos = append(videos, videoItem)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}

	stderrWg.Wait()

	if err := cmd.Wait(); err != nil {
		log.Printf("yt-dlp command failed: %v", err)
		log.Printf("stderr output: %v", stderrLines)
	}

	if sm.ClearAndCheckCanceled() {
		return nil, nil, 0, "", true
	}

	return videos, stderrLines, skippedLiveShort, "", false
}

func executeYTDLP(sm *SearchManager, searchURL string, searchLimit int, cookiesBrowser, cookiesFile string) any {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config, using defaults: %v", err)
		cfg = config.GetDefault()
	}

	ytDlpPath := cfg.YTDLPPath
	if ytDlpPath == "" {
		ytDlpPath = "yt-dlp"
	}

	if err := exec.Command(ytDlpPath, "--version").Run(); err != nil {
		if err.Error() == "exec: \""+ytDlpPath+"\": executable file not found in $PATH" ||
			strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(err.Error(), "no such file or directory") {
			errMsg := "yt-dlp not found. Please install yt-dlp: https://github.com/yt-dlp/yt-dlp#installation"
			return types.SearchResultMsg{Err: errMsg}
		}

		errMsg := fmt.Sprintf("Failed to run yt-dlp: %v\nPlease check your yt-dlp installation", err)
		return types.SearchResultMsg{Err: errMsg}
	}

	if cookiesBrowser == "" {
		cookiesBrowser = cfg.CookiesBrowser
	}
	if cookiesFile == "" {
		cookiesFile = cfg.CookiesFile
	}

	var args []string
	if cookiesBrowser != "" {
		args = append(args, "--cookies-from-browser", cookiesBrowser)
	} else if cookiesFile != "" {
		args = append(args, "--cookies", cookiesFile)
	}

	targetLimit := searchLimit
	fetchLimit := searchLimit
	var videos []list.Item
	var stderrLines []string

	for attempts := 0; attempts < 4; attempts++ {
		var skippedLiveShort int
		var errMsg string
		var canceled bool

		videos, stderrLines, skippedLiveShort, errMsg, canceled = runYTDLPCommand(sm, ytDlpPath, searchURL, fetchLimit, args)
		if canceled {
			return nil
		}

		if errMsg != "" {
			return types.SearchResultMsg{Err: errMsg}
		}

		if len(videos) >= targetLimit {
			return types.SearchResultMsg{Videos: videos[:targetLimit]}
		}

		if skippedLiveShort == 0 {
			break
		}

		nextLimit := targetLimit + skippedLiveShort
		if nextLimit <= fetchLimit {
			break
		}

		fetchLimit = nextLimit
	}

	var errMsg string
	if len(videos) == 0 {
		for _, line := range stderrLines {
			if strings.Contains(line, "[Errno 101]") || strings.Contains(line, "[Errno -3]") {
				errMsg = "Please Check Your Internet connection"
			} else if strings.Contains(line, "HTTP Error 404") || strings.Contains(line, "Requested entity was not found") {
				if strings.Contains(searchURL, "/playlist?list=") {
					errMsg = "Playlist not found"
				} else {
					errMsg = "Channel not found"
				}
			} else if strings.Contains(line, "Private playlist") || strings.Contains(line, "This playlist is private") {
				errMsg = "This playlist is private"
			} else if strings.Contains(line, "Playlist does not exist") {
				errMsg = "Playlist does not exist"
			}
		}

		return types.SearchResultMsg{Err: errMsg}
	} else {
		return types.SearchResultMsg{Videos: videos}
	}
}

func PerformSearch(sm *SearchManager, query, sortParam string, searchLimit int, cookiesBrowser, cookiesFile string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		query = strings.TrimSpace(query)

		urlType, url := ParseSearchQuery(query)
		if urlType == "video" {
			return types.StartFormatMsg{URL: url}
		}

		log.Print("urlType: ", urlType)
		return executeYTDLP(sm, url, searchLimit, cookiesBrowser, cookiesFile)
	})
}

func PerformChannelSearch(sm *SearchManager, input string, searchLimit int, cookiesBrowser, cookiesFile string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		channelURL := BuildChannelURL(input)
		return executeYTDLP(sm, channelURL, searchLimit, cookiesBrowser, cookiesFile)
	})
}

func PerformPlaylistSearch(sm *SearchManager, query string, searchLimit int, cookiesBrowser, cookiesFile string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		playlistURL := BuildPlaylistURL(query)
		return executeYTDLP(sm, playlistURL, searchLimit, cookiesBrowser, cookiesFile)
	})
}

func CancelSearch(sm *SearchManager) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if err := sm.Cancel(); err != nil {
			log.Printf("Failed to cancel search: %v", err)
		}

		return types.CancelSearchMsg{}
	})
}
