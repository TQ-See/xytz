package types

import "github.com/charmbracelet/bubbles/list"

type State string

const (
	StateSearchInput = "search_input"
	StateLoading     = "loading"
	StateVideoList   = "video_list"
)

type StartSearchMsg struct {
	Query string
}

type VideoItem struct {
	ID         string
	VideoTitle string
	Desc       string
	Views      float64
	Duration   float64
}

func (i VideoItem) Title() string       { return i.VideoTitle }
func (i VideoItem) Description() string { return i.Desc }
func (i VideoItem) FilterValue() string { return i.VideoTitle }

type SearchResultMsg struct {
	Videos []list.Item
	Err    string
}
