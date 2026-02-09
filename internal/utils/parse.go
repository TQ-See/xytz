package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/xdagiz/xytz/internal/types"
)

func extractAfterDelimiter(s, delimiter string, trailingDelimiters ...string) string {
	parts := strings.Split(s, delimiter)
	if len(parts) <= 1 {
		return ""
	}

	result := parts[1]
	for _, delim := range trailingDelimiters {
		if idx := strings.Index(result, delim); idx != -1 {
			result = result[:idx]
		}
	}

	return result
}

func ExtractVideoID(url string) string {
	if strings.Contains(url, "youtube.com/watch") && strings.Contains(url, "v=") {
		if result := extractAfterDelimiter(url, "v=", "&", "#"); result != "" {
			return result
		}
	}

	if strings.Contains(url, "youtu.be/") {
		if result := extractAfterDelimiter(url, "youtu.be/", "&", "#", "?"); result != "" {
			return result
		}
	}

	if strings.Contains(url, "youtube.com/embed/") {
		if result := extractAfterDelimiter(url, "youtube.com/embed/", "&", "#"); result != "" {
			return result
		}
	}

	return ""
}

func ExtractChannelUsername(input string) string {
	input = strings.TrimSpace(input)

	if after, ok := strings.CutPrefix(input, "@"); ok {
		return after
	}

	if strings.Contains(input, "youtube.com/@") {
		if result := extractAfterDelimiter(input, "@", "/"); result != "" {
			return result
		}
	}

	if strings.Contains(input, "/channel/") {
		if result := extractAfterDelimiter(input, "/channel/", "?"); result != "" {
			return result
		}
	}

	if strings.Contains(input, "/c/") {
		if result := extractAfterDelimiter(input, "/c/", "/"); result != "" {
			return result
		}
	}

	return input
}

func ParseVideoItem(line string) (types.VideoItem, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return types.VideoItem{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if data == nil {
		return types.VideoItem{}, fmt.Errorf("received nil data")
	}

	title, ok := data["title"].(string)
	if !ok || title == "" {
		return types.VideoItem{}, fmt.Errorf("missing title in video data")
	}
	videoID, ok := data["id"].(string)
	if !ok || videoID == "" {
		return types.VideoItem{}, fmt.Errorf("missing video ID in video data")
	}

	channel, ok := data["uploader"].(string)
	if !ok || channel == "" {
		if playlistUploader, ok := data["playlist_uploader"].(string); ok && playlistUploader != "" {
			channel = playlistUploader
		}
	}

	var viewCountFloat float64
	if vc, ok := data["view_count"]; ok {
		viewCountFloat = parseFloat(vc)
	}

	var durationFloat float64
	if d, ok := data["duration"]; ok {
		durationFloat = parseFloat(d)
	}

	if durationFloat == 0 {
		return types.VideoItem{}, fmt.Errorf("skipping live/short content with zero duration")
	}

	viewsStr := FormatNumber(viewCountFloat)
	durationStr := FormatDuration(durationFloat)

	channelLen := len(channel)
	if channelLen > 30 {
		channel = channel[:27] + "..."
	}

	desc := fmt.Sprintf("%s • %s views • %s", durationStr, viewsStr, channel)

	videoItem := types.VideoItem{
		ID:         videoID,
		VideoTitle: title,
		Desc:       desc,
		Views:      viewCountFloat,
		Duration:   durationFloat,
		Channel:    channel,
	}

	return videoItem, nil
}

func parseFloat(v any) float64 {
	switch val := v.(type) {
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case float64:
		return val
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	default:
		s := fmt.Sprintf("%v", v)
		if s != "" {
			f, _ := strconv.ParseFloat(s, 64)
			return f
		}
	}

	return 0
}
