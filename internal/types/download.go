package types

import tea "github.com/charmbracelet/bubbletea"

type DownloadOption struct {
	Name           string
	KeyBinding     tea.KeyType
	ConfigField    string
	RequiresFFmpeg bool
	Enabled        bool
}

func DownloadOptions() []DownloadOption {
	return []DownloadOption{
		{
			Name:           "Add Subtitles",
			KeyBinding:     tea.KeyCtrlS,
			ConfigField:    "EmbedSubtitles",
			RequiresFFmpeg: true,
		},
		{
			Name:           "Add Metadata",
			KeyBinding:     tea.KeyCtrlJ,
			ConfigField:    "EmbedMetadata",
			RequiresFFmpeg: true,
		},
		{
			Name:           "Add Chapters",
			KeyBinding:     tea.KeyCtrlL,
			ConfigField:    "EmbedChapters",
			RequiresFFmpeg: true,
		},
	}
}

type DownloadRequest struct {
	URL      string
	FormatID string

	IsAudioTab bool
	ABR        float64

	Title           string
	QueueIndex      int
	QueueTotal      int
	URLs            []string
	Videos          []VideoItem
	UnfinishedKey   string
	UnfinishedTitle string
	UnfinishedDesc  string

	Options []DownloadOption

	CookiesFromBrowser string
	Cookies            string
}
