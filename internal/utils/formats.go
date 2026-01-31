package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/types"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func formatQuality(resolution string) string {
	if resolution == "" || resolution == "?" {
		return resolution
	}

	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return resolution
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return resolution
	}

	switch {
	case height >= 4320:
		return "8k"
	case height >= 2160:
		return "4k"
	case height >= 1440:
		return "2k"
	case height >= 1080:
		return "1080p"
	case height >= 720:
		return "720p"
	case height >= 480:
		return "480p"
	case height >= 360:
		return "360p"
	case height >= 240:
		return "240p"
	case height >= 144:
		return "144p"
	default:
		return resolution
	}
}

func FetchFormats(url string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.GetDefault()
		}
		ytDlpPath := cfg.YTDLPPath
		if ytDlpPath == "" {
			ytDlpPath = "yt-dlp"
		}
		cmd := exec.Command(ytDlpPath, "-J", url)
		out, err := cmd.Output()
		if err != nil {
			errMsg := fmt.Sprintf("Format fetch error: %v", err)
			return types.SearchResultMsg{Err: errMsg}
		}
		var data map[string]any
		if err := json.Unmarshal(out, &data); err != nil {
			errMsg := fmt.Sprintf("JSON parse error: %v", err)
			return types.SearchResultMsg{Err: errMsg}
		}

		formatsAny, _ := data["formats"].([]any)
		var videoFormats []list.Item
		var audioFormats []list.Item
		var thumbnailFormats []list.Item
		var allFormats []list.Item

		audioLanguages := make(map[string]bool)
		for _, fAny := range formatsAny {
			f, ok := fAny.(map[string]any)
			if !ok {
				continue
			}

			acodec, _ := f["acodec"].(string)
			if acodec != "none" && acodec != "" {
				lang, _ := f["language"].(string)
				if lang == "" {
					lang, _ = f["lang"].(string)
				}
				if lang != "" && lang != "und" {
					audioLanguages[lang] = true
				}
			}
		}

		showLanguage := len(audioLanguages) > 1

		for _, fAny := range formatsAny {
			f, ok := fAny.(map[string]any)
			if !ok {
				continue
			}

			formatID, _ := f["format_id"].(string)
			ext, _ := f["ext"].(string)
			resolution, _ := f["resolution"].(string)
			acodec, _ := f["acodec"].(string)
			vcodec, _ := f["vcodec"].(string)
			abr, _ := f["abr"].(float64)
			fps, _ := f["fps"].(float64)
			tbr, _ := f["tbr"].(float64)

			if formatID == "" {
				continue
			}

			if ext == "" {
				continue
			}

			if resolution == "" || resolution == "Unknown" {
				resolution = "?"
			}

			formatType := ""
			isVideoAudio := false
			isAudioOnly := false
			isThumbnail := ext == "mhtml"

			if vcodec != "none" && vcodec != "" {
				if acodec != "none" && acodec != "" {
					formatType = "video+audio"
					isVideoAudio = true
				} else {
					formatType = "video-only"
				}
			} else if acodec != "none" && acodec != "" {
				formatType = "audio-only"
				isAudioOnly = true
			} else if isThumbnail {
				formatType = "thumbnail"
			} else {
				formatType = "unknown"
			}

			size, _ := f["filesize"].(float64)
			sizeApprox, _ := f["filesize_approx"].(float64)
			if size == 0 {
				size = sizeApprox
			}
			sizeStr := bytesToHuman(size)

			lang := ""
			if showLanguage {
				lang, _ = f["language"].(string)
				if lang == "" {
					lang, _ = f["lang"].(string)
				}
				if lang == "" || lang == "und" {
					lang = "unknown"
				}
			}

			title := ext
			if isAudioOnly {
				if abr > 0 {
					title = fmt.Sprintf("%s @%dk", ext, int(abr))
				}
			} else if isThumbnail {
				title = formatQuality(resolution)
			} else {
				quality := formatQuality(resolution)
				if fps > 0 {
					quality = fmt.Sprintf("%s%.0f", quality, fps)
				}
				title = quality
				if tbr > 0 {
					title = fmt.Sprintf("%s @%s", title, formatBitrate(tbr))
				}
				title = fmt.Sprintf("%s %s", title, ext)
			}

			if showLanguage && (acodec != "none" && acodec != "") {
				title = fmt.Sprintf("%s [%s]", title, lang)
			}

			formatItem := types.FormatItem{
				FormatTitle: title,
				FormatValue: formatID,
				Size:        sizeStr,
				Language:    lang,
				Resolution:  resolution,
				FormatType:  formatType,
			}

			allFormats = append(allFormats, formatItem)

			if isVideoAudio {
				// Filter out formats below 360p (144p and 240p)
				if !strings.Contains(title, "144p") && !strings.Contains(title, "240p") {
					videoFormats = append(videoFormats, formatItem)
				}
			} else if isAudioOnly {
				audioFormats = append(audioFormats, formatItem)
			} else if isThumbnail {
				thumbnailFormats = append(thumbnailFormats, formatItem)
			}
		}

		return types.FormatResultMsg{
			VideoFormats:     videoFormats,
			AudioFormats:     audioFormats,
			ThumbnailFormats: thumbnailFormats,
			AllFormats:       allFormats,
		}
	})
}
