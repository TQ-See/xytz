package models

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"
)

func setupModelTestEnv(t *testing.T) {
	t.Helper()

	origConfigDir := config.GetConfigDir
	origUnfinishedPath := utils.GetUnfinishedFilePath
	origHistoryPath := utils.GetHistoryFilePath

	tmpDir := t.TempDir()
	config.GetConfigDir = func() string {
		return filepath.Join(tmpDir, "config")
	}
	utils.GetUnfinishedFilePath = func() string {
		return filepath.Join(tmpDir, "unfinished.json")
	}
	utils.GetHistoryFilePath = func() string {
		return filepath.Join(tmpDir, "history")
	}

	t.Cleanup(func() {
		config.GetConfigDir = origConfigDir
		utils.GetUnfinishedFilePath = origUnfinishedPath
		utils.GetHistoryFilePath = origHistoryPath
	})
}

func cmdMsg(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected non-nil command")
	}

	return cmd()
}

func TestSearchModelEnterEmptyQueryShowsError(t *testing.T) {
	setupModelTestEnv(t)

	m := NewSearchModel()
	m.Input.SetValue("")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.ErrMsg != "Please enter a query or URL" {
		t.Fatalf("ErrMsg = %q, want %q", m.ErrMsg, "Please enter a query or URL")
	}
}

func TestSearchModelSlashHelpTogglesAndClearsInput(t *testing.T) {
	setupModelTestEnv(t)

	m := NewSearchModel()
	m.Input.SetValue("/help")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if !m.Help.Visible {
		t.Fatalf("expected help to be visible")
	}
	if m.Input.Value() != "" {
		t.Fatalf("input value = %q, want empty", m.Input.Value())
	}
}

func TestSearchModelSlashChannelReturnsStartChannelMsg(t *testing.T) {
	setupModelTestEnv(t)

	m := NewSearchModel()
	m.Input.SetValue("/channel @xdagiz")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	msg := cmdMsg(t, cmd)
	got, ok := msg.(types.StartChannelURLMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartChannelURLMsg", msg)
	}
	if got.ChannelName != "xdagiz" {
		t.Fatalf("ChannelName = %q, want xdagiz", got.ChannelName)
	}
}

func TestSearchModelResumeSlashAndEnterStartsResumeDownload(t *testing.T) {
	setupModelTestEnv(t)

	err := utils.SaveUnfinished([]utils.UnfinishedDownload{
		{
			URL:       "queue:test",
			URLs:      []string{"https://example.com/v1"},
			Videos:    []types.VideoItem{{ID: "v1", VideoTitle: "Video 1"}},
			FormatID:  "best",
			Title:     "Queued downloads",
			Desc:      "1 item left",
			Timestamp: time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("SaveUnfinished error: %v", err)
	}

	m := NewSearchModel()
	m.Input.SetValue("/resume")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	if cmd != nil {
		t.Fatalf("expected nil command when opening resume list")
	}
	if !m.ResumeList.Visible {
		t.Fatalf("expected resume list to be visible")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated
	msg := cmdMsg(t, cmd)
	resumeMsg, ok := msg.(types.StartResumeDownloadMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartResumeDownloadMsg", msg)
	}
	if resumeMsg.FormatID != "best" {
		t.Fatalf("FormatID = %q, want best", resumeMsg.FormatID)
	}
	if len(resumeMsg.URLs) != 1 || resumeMsg.URLs[0] != "https://example.com/v1" {
		t.Fatalf("URLs = %#v, want one expected URL", resumeMsg.URLs)
	}
}

func TestSearchModelResumeEscHidesList(t *testing.T) {
	setupModelTestEnv(t)

	m := NewSearchModel()
	m.ResumeList.Visible = true
	m.Input.SetValue("abc")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated

	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.ResumeList.Visible {
		t.Fatalf("expected resume list to be hidden after esc")
	}
	if m.Input.Value() != "" {
		t.Fatalf("input = %q, want empty", m.Input.Value())
	}
}

func TestVideoListSpaceTogglesSelection(t *testing.T) {
	setupModelTestEnv(t)

	m := NewVideoListModel()
	m.SetItems([]list.Item{types.VideoItem{ID: "a", VideoTitle: "Video A"}})
	m.List.Select(0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated
	if len(m.SelectedVideos) != 1 || m.SelectedVideos[0].ID != "a" {
		t.Fatalf("selected after first space = %#v, want one selected video", m.SelectedVideos)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated
	if len(m.SelectedVideos) != 0 {
		t.Fatalf("selected after second space = %#v, want empty", m.SelectedVideos)
	}
}

func TestVideoListEnterWithSelectedVideosReturnsQueueConfirm(t *testing.T) {
	setupModelTestEnv(t)

	m := NewVideoListModel()
	m.SetItems([]list.Item{
		types.VideoItem{ID: "a", VideoTitle: "Video A"},
		types.VideoItem{ID: "b", VideoTitle: "Video B"},
	})
	m.SelectedVideos = []types.VideoItem{
		{ID: "a", VideoTitle: "Video A"},
		{ID: "b", VideoTitle: "Video B"},
	}
	m.List.Select(0)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	msg := cmdMsg(t, cmd)
	got, ok := msg.(types.StartQueueConfirmMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartQueueConfirmMsg", msg)
	}
	if len(got.Videos) != 2 {
		t.Fatalf("queue confirm videos len = %d, want 2", len(got.Videos))
	}
}

func TestVideoListDWithSelectedVideosReturnsQueueDownload(t *testing.T) {
	setupModelTestEnv(t)

	m := NewVideoListModel()
	m.SetItems([]list.Item{
		types.VideoItem{ID: "a", VideoTitle: "Video A"},
		types.VideoItem{ID: "b", VideoTitle: "Video B"},
	})
	m.SelectedVideos = []types.VideoItem{
		{ID: "a", VideoTitle: "Video A"},
		{ID: "b", VideoTitle: "Video B"},
	}
	m.List.Select(0)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated

	msg := cmdMsg(t, cmd)
	got, ok := msg.(types.StartQueueDownloadMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.StartQueueDownloadMsg", msg)
	}
	if len(got.Videos) != 2 {
		t.Fatalf("queue download videos len = %d, want 2", len(got.Videos))
	}
}

func TestVideoListEnterWithErrorReturnsBackMessage(t *testing.T) {
	setupModelTestEnv(t)

	m := NewVideoListModel()
	m.ErrMsg = "Channel not found"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	msg := cmdMsg(t, cmd)
	if _, ok := msg.(types.BackFromVideoListMsg); !ok {
		t.Fatalf("cmd msg type = %T, want types.BackFromVideoListMsg", msg)
	}
}

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
