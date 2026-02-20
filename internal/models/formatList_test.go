package models

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdagiz/xytz/internal/types"
)

func TestFormatListTabCycleAndReverse(t *testing.T) {
	setupModelTestEnv(t)

	m := NewFormatListModel()
	m.SetFormats(
		[]list.Item{types.FormatItem{FormatTitle: "V", FormatValue: "137"}},
		[]list.Item{types.FormatItem{FormatTitle: "A", FormatValue: "140"}},
		[]list.Item{types.FormatItem{FormatTitle: "T", FormatValue: "sb0"}},
		nil,
	)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated
	if m.ActiveTab != FormatTabAudio {
		t.Fatalf("tab from video => %v, want audio", m.ActiveTab)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated
	if m.ActiveTab != FormatTabVideo {
		t.Fatalf("shift+tab from audio => %v, want video", m.ActiveTab)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated
	if m.ActiveTab != FormatTabCustom {
		t.Fatalf("shift+tab from video => %v, want custom", m.ActiveTab)
	}
}

func TestFormatListEnterOnSelectedVideoFormatReturnsStartDownload(t *testing.T) {
	setupModelTestEnv(t)

	m := NewFormatListModel()
	m.URL = "https://www.youtube.com/watch?v=abc"
	m.SetFormats(
		[]list.Item{types.FormatItem{FormatTitle: "1080p", FormatValue: "137+140"}},
		nil,
		nil,
		nil,
	)
	m.List.Select(0)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	msg := cmdMsg(t, cmd)
	got, ok := msg.(types.StartDownloadMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartDownloadMsg", msg)
	}
	if got.FormatID != "137+140" {
		t.Fatalf("FormatID = %q, want 137+140", got.FormatID)
	}
}

func TestFormatListCustomAutocompleteTabReplacesToken(t *testing.T) {
	setupModelTestEnv(t)

	m := NewFormatListModel()
	m.ActiveTab = FormatTabCustom
	m.AllFormats = []list.Item{
		types.FormatItem{FormatTitle: "1080p", FormatValue: "137"},
	}
	m.CustomInput.SetValue("best+13")
	m.Autocomplete.Show("best+13", m.AllFormats)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated

	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.CustomInput.Value() != "best+137" {
		t.Fatalf("custom input = %q, want best+137", m.CustomInput.Value())
	}
	if m.Autocomplete.Visible {
		t.Fatalf("autocomplete should be hidden after selection")
	}
}

func TestFormatListCustomEnterQueueReturnsStartQueueDownload(t *testing.T) {
	setupModelTestEnv(t)

	m := NewFormatListModel()
	m.ActiveTab = FormatTabCustom
	m.IsQueue = true
	m.QueueVideos = []types.VideoItem{
		{ID: "a", VideoTitle: "Video A"},
		{ID: "b", VideoTitle: "Video B"},
	}
	m.CustomInput.SetValue("bestvideo+bestaudio")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	msg := cmdMsg(t, cmd)
	got, ok := msg.(types.StartQueueDownloadMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartQueueDownloadMsg", msg)
	}
	if got.FormatID != "bestvideo+bestaudio" {
		t.Fatalf("FormatID = %q, want bestvideo+bestaudio", got.FormatID)
	}
	if len(got.Videos) != 2 {
		t.Fatalf("Videos len = %d, want 2", len(got.Videos))
	}
}
