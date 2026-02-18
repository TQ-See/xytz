package app

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	zone "github.com/lrstanley/bubblezone"
	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/models"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"
)

type noInitModel struct {
	*Model
}

func (m noInitModel) Init() tea.Cmd {
	return nil
}

func setupAppTeaEnv(t *testing.T) {
	t.Helper()

	origConfigDir := config.GetConfigDir
	origUnfinishedPath := utils.GetUnfinishedFilePath

	tmpDir := t.TempDir()
	config.GetConfigDir = func() string {
		return filepath.Join(tmpDir, "config")
	}
	utils.GetUnfinishedFilePath = func() string {
		return filepath.Join(tmpDir, "unfinished.json")
	}

	t.Cleanup(func() {
		config.GetConfigDir = origConfigDir
		utils.GetUnfinishedFilePath = origUnfinishedPath
	})
}

func newAppTeaModel(t *testing.T, setup func(m *Model)) (*Model, *teatest.TestModel) {
	t.Helper()
	setupAppTeaEnv(t)

	zone.NewGlobal()
	t.Cleanup(zone.Close)

	m := NewModel()
	m.Width = 120
	m.Height = 40
	if setup != nil {
		setup(m)
	}

	tm := teatest.NewTestModel(t, noInitModel{Model: m}, teatest.WithInitialTermSize(120, 40))
	m.Program = tm.GetProgram()

	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	return m, tm
}

func waitForState(t *testing.T, m *Model, want types.State) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if m.State == want {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("state did not reach %q, got %q", want, m.State)
}

func waitForViewContains(t *testing.T, m *Model, s string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(m.View(), s) {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("view did not contain %q; got:\n%s", s, m.View())
}

func TestAppTeaStateSearchInputView(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateSearchInput
	})

	waitForViewContains(t, m, "Sort By")
	waitForViewContains(t, m, "Download Options")
}

func TestAppTeaStateLoadingViewByType(t *testing.T) {
	tests := []struct {
		name        string
		loadingType string
		query       string
		channel     string
		want        string
	}{
		{name: "search", loadingType: "search", query: "golang", want: "Searching for"},
		{name: "format", loadingType: "format", want: "Loading formats..."},
		{name: "channel", loadingType: "channel", channel: "xdagiz", want: "Loading videos for channel"},
		{name: "playlist", loadingType: "playlist", query: "my-playlist", want: "Searching playlist:"},
		{name: "queue", loadingType: "queue", want: "Starting queue download..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := newAppTeaModel(t, func(m *Model) {
				m.State = types.StateLoading
				m.LoadingType = tt.loadingType
				m.CurrentQuery = tt.query
				m.VideoList.ChannelName = tt.channel
			})

			waitForViewContains(t, m, tt.want)
		})
	}
}

func TestAppTeaStateVideoListView(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateVideoList
		m.VideoList.CurrentQuery = "lofi"
		m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "abc", VideoTitle: "Lofi Mix"}})
	})

	waitForViewContains(t, m, "Search Results for: lofi")
	waitForViewContains(t, m, "Lofi Mix")
}

func TestAppTeaStateFormatListView(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateFormatList
		m.FormatList.URL = utils.BuildVideoURL("abc")
		m.FormatList.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Video A"}
		m.FormatList.ShowVideoInfo = true
		m.FormatList.SetFormats(
			[]list.Item{types.FormatItem{FormatTitle: "1080p", FormatValue: "137+140"}},
			nil,
			nil,
			nil,
		)
	})

	waitForViewContains(t, m, "Select a Format")
	waitForViewContains(t, m, "Video A")
	waitForViewContains(t, m, "1080p")
}

func TestAppTeaStateDownloadView(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Download Me"}
		m.Download.Phase = "[download]"
	})

	waitForViewContains(t, m, "Download Me")
	waitForViewContains(t, m, "Downloading")
}

func TestAppTeaTransitionCancelSearchToSearchInput(t *testing.T) {
	m, tm := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateLoading
		m.LoadingType = "search"
		m.CurrentQuery = "abc"
	})

	tm.Send(types.CancelSearchMsg{})
	waitForState(t, m, types.StateSearchInput)
	waitForViewContains(t, m, "Sort By")

	if m.State != types.StateSearchInput {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
	}
}

func TestAppTeaTransitionCancelFormatsToVideoList(t *testing.T) {
	m, tm := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateFormatList
		m.VideoList.CurrentQuery = "abc"
		m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "abc", VideoTitle: "A"}})
	})

	tm.Send(types.CancelFormatsMsg{})
	waitForState(t, m, types.StateVideoList)
	waitForViewContains(t, m, "Search Results for: abc")

	if m.State != types.StateVideoList {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateVideoList)
	}
}

func TestAppTeaTransitionBackFromVideoListToSearchInput(t *testing.T) {
	m, tm := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateVideoList
		m.VideoList.CurrentQuery = "abc"
		m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "abc", VideoTitle: "A"}})
	})

	tm.Send(types.BackFromVideoListMsg{})
	waitForState(t, m, types.StateSearchInput)
	waitForViewContains(t, m, "Sort By")

	if m.State != types.StateSearchInput {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
	}
}

func TestAppTeaTransitionDownloadCompleteToSearchInput(t *testing.T) {
	m, tm := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "A"}
		m.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "A"}
	})

	tm.Send(types.DownloadCompleteMsg{})
	waitForState(t, m, types.StateSearchInput)
	waitForViewContains(t, m, "Sort By")

	if m.State != types.StateSearchInput {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
	}
	if m.SelectedVideo.ID != "" {
		t.Fatalf("m.SelectedVideo.ID = %q, want empty", m.SelectedVideo.ID)
	}
}

func TestAppTeaTransitionDownloadBackKeyWhenCompleted(t *testing.T) {
	m, tm := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.Completed = true
		m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "A"}
		m.FormatList.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "A"}
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	waitForState(t, m, types.StateFormatList)
	waitForViewContains(t, m, "Select a Format")

	if m.State != types.StateFormatList {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateFormatList)
	}
}

func TestAppTeaStatusBarKeysByState(t *testing.T) {
	t.Run("search input", func(t *testing.T) {
		m, _ := newAppTeaModel(t, func(m *Model) {
			m.State = types.StateSearchInput
		})

		waitForViewContains(t, m, "Ctrl+c")
		waitForViewContains(t, m, "quit")
	})

	t.Run("loading", func(t *testing.T) {
		m, _ := newAppTeaModel(t, func(m *Model) {
			m.State = types.StateLoading
			m.LoadingType = "search"
		})

		waitForViewContains(t, m, "Esc/c")
		waitForViewContains(t, m, "cancel")
	})

	t.Run("video list", func(t *testing.T) {
		m, _ := newAppTeaModel(t, func(m *Model) {
			m.State = types.StateVideoList
			m.VideoList.CurrentQuery = "abc"
			m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "abc", VideoTitle: "Video A"}})
		})

		waitForViewContains(t, m, "d")
		waitForViewContains(t, m, "Download")
		waitForViewContains(t, m, "Space")
		waitForViewContains(t, m, "select")
	})

	t.Run("download active", func(t *testing.T) {
		m, _ := newAppTeaModel(t, func(m *Model) {
			m.State = types.StateDownload
			m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Video A"}
		})

		waitForViewContains(t, m, "p/space")
		waitForViewContains(t, m, "pause")
		waitForViewContains(t, m, "Esc/c")
		waitForViewContains(t, m, "cancel")
	})

	t.Run("download completed", func(t *testing.T) {
		m, _ := newAppTeaModel(t, func(m *Model) {
			m.State = types.StateDownload
			m.Download.Completed = true
			m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Video A"}
		})

		waitForViewContains(t, m, "Enter")
		waitForViewContains(t, m, "back to search")
	})
}

func TestAppTeaQueueSummaryConsistencyCompleted(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.IsQueue = true
		m.Download.Completed = true
		m.Download.QueueIndex = 3
		m.Download.QueueTotal = 3
		m.Download.QueueItems = []types.QueueItem{
			{Index: 1, Video: types.VideoItem{ID: "a", VideoTitle: "A"}, Status: types.QueueStatusComplete},
			{Index: 2, Video: types.VideoItem{ID: "b", VideoTitle: "B"}, Status: types.QueueStatusError, Error: "boom"},
			{Index: 3, Video: types.VideoItem{ID: "c", VideoTitle: "C"}, Status: types.QueueStatusSkipped},
		}
	})

	waitForViewContains(t, m, "Queue Summary:")
	waitForViewContains(t, m, "1 complete | 1 failed | 1 skipped")
	waitForViewContains(t, m, "A")
	waitForViewContains(t, m, "B")
	waitForViewContains(t, m, "C")
}

func TestAppTeaQueueSummaryConsistencyCancelled(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.IsQueue = true
		m.Download.Cancelled = true
		m.Download.QueueIndex = 3
		m.Download.QueueTotal = 3
		m.Download.QueueItems = []types.QueueItem{
			{Index: 1, Video: types.VideoItem{ID: "a", VideoTitle: "A"}, Status: types.QueueStatusComplete},
			{Index: 2, Video: types.VideoItem{ID: "b", VideoTitle: "B"}, Status: types.QueueStatusError, Error: "boom"},
			{Index: 3, Video: types.VideoItem{ID: "c", VideoTitle: "C"}, Status: types.QueueStatusSkipped},
		}
	})

	waitForViewContains(t, m, "Queue Cancelled:")
	waitForViewContains(t, m, "1 complete | 1 failed | 1 skipped")
}

func TestAppTeaQueueErrorScreenShowsActions(t *testing.T) {
	m, _ := newAppTeaModel(t, func(m *Model) {
		m.State = types.StateDownload
		m.Download.IsQueue = true
		m.Download.QueueError = "network down"
		m.Download.QueueIndex = 2
		m.Download.QueueTotal = 2
		m.Download.QueueItems = []types.QueueItem{
			{Index: 1, Video: types.VideoItem{ID: "a", VideoTitle: "A"}, Status: types.QueueStatusComplete},
			{Index: 2, Video: types.VideoItem{ID: "b", VideoTitle: "B"}, Status: types.QueueStatusError, Error: "network down"},
		}
	})

	waitForViewContains(t, m, "Error: network down")
	waitForViewContains(t, m, "[s] Skip")
	waitForViewContains(t, m, "[r] Retry")
	waitForViewContains(t, m, "[c/esc] Cancel queue")
}

func TestAppEscInLoadingSearchTriggersCancelSearch(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateLoading
	m.LoadingType = "search"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if cmd == nil {
		t.Fatalf("expected non-nil cancel cmd")
	}

	msg := cmd()
	cancelMsg, ok := msg.(types.CancelSearchMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.CancelSearchMsg", msg)
	}

	updated, _ = m.Update(cancelMsg)
	m = updated.(*Model)
	if m.State != types.StateSearchInput {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
	}
}

func TestAppEscInLoadingFormatTriggersCancelFormats(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateLoading
	m.LoadingType = "format"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if cmd == nil {
		t.Fatalf("expected non-nil cancel cmd")
	}

	msg := cmd()
	cancelMsg, ok := msg.(types.CancelFormatsMsg)
	if !ok {
		t.Fatalf("cmd msg type = %T, want types.CancelFormatsMsg", msg)
	}

	updated, _ = m.Update(cancelMsg)
	m = updated.(*Model)
	if m.State != types.StateVideoList {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateVideoList)
	}
}

func TestAppEscInVideoListClearsSelectionFirst(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateVideoList
	m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "a", VideoTitle: "A"}})
	m.VideoList.SelectedVideos = []types.VideoItem{{ID: "a", VideoTitle: "A"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if cmd != nil {
		t.Fatalf("expected nil cmd when clearing selection")
	}
	if m.State != types.StateVideoList {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateVideoList)
	}
	if len(m.VideoList.SelectedVideos) != 0 {
		t.Fatalf("SelectedVideos len = %d, want 0", len(m.VideoList.SelectedVideos))
	}
}

func TestAppEscInVideoListBacksToSearchWhenNotFiltering(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateVideoList
	m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "a", VideoTitle: "A"}})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if cmd != nil {
		t.Fatalf("expected nil cmd")
	}
	if m.State != types.StateSearchInput {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
	}
}

func TestAppEscInVideoListWhileFilteringStaysInVideoList(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateVideoList
	m.VideoList.SetItems([]list.Item{types.VideoItem{ID: "a", VideoTitle: "A"}})
	m.VideoList.List.SetFilterState(list.Filtering)
	m.VideoList.List.FilterInput.SetValue("a")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if cmd != nil {
		t.Fatalf("expected nil cmd")
	}
	if m.State != types.StateVideoList {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateVideoList)
	}
	if m.VideoList.List.FilterState() != list.Unfiltered {
		t.Fatalf("filter state = %v, want %v", m.VideoList.List.FilterState(), list.Unfiltered)
	}
}

func TestAppEscInFormatListBackBehavior(t *testing.T) {
	setupAppTeaEnv(t)

	t.Run("no selected video goes to search input", func(t *testing.T) {
		m := NewModel()
		m.State = types.StateFormatList
		m.FormatList.ActiveTab = 0

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(*Model)
		if cmd != nil {
			t.Fatalf("expected nil cmd")
		}
		if m.State != types.StateSearchInput {
			t.Fatalf("m.State = %q, want %q", m.State, types.StateSearchInput)
		}
	})

	t.Run("selected video goes to video list", func(t *testing.T) {
		m := NewModel()
		m.State = types.StateFormatList
		m.SelectedVideo = types.VideoItem{ID: "a", VideoTitle: "A"}
		m.FormatList.ActiveTab = 0

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(*Model)
		if cmd != nil {
			t.Fatalf("expected nil cmd")
		}
		if m.State != types.StateVideoList {
			t.Fatalf("m.State = %q, want %q", m.State, types.StateVideoList)
		}
	})
}

func TestAppEscInSearchInputHidesHelp(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateSearchInput
	m.Search.Help.Visible = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if m.Search.Help.Visible {
		t.Fatalf("expected help to be hidden after esc")
	}
}

func TestPlayVideoMsgTriggersMPVWithDefaultQuality(t *testing.T) {
	setupAppTeaEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	cfg.DefaultQuality = "720p"
	if err := cfg.Save(); err != nil {
		t.Fatalf("cfg.Save() error: %v", err)
	}

	m := NewModel()
	updated, cmd := m.Update(types.PlayVideoMsg{SelectedVideo: types.VideoItem{ID: "abc123"}})
	m = updated.(*Model)

	if cmd == nil {
		t.Fatalf("expected non-nil command")
	}

	updated, _ = m.Update(cmd())
	m = updated.(*Model)

	if m == nil {
		t.Fatalf("expected model")
	}
	if m.State != types.StateVideoPlaying {
		t.Fatalf("expected StateVideoPlaying, got %s", m.State)
	}
}

func TestPlayVideoMsgPassesCorrectFormatToMPV(t *testing.T) {
	setupAppTeaEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	cfg.DefaultQuality = "720p"
	if err := cfg.Save(); err != nil {
		t.Fatalf("cfg.Save() error: %v", err)
	}

	m := NewModel()
	videoURL := utils.BuildVideoURL("abc123")

	_, _ = m.Update(types.PlayVideoMsg{SelectedVideo: types.VideoItem{ID: "abc123", VideoTitle: "Test Video"}})

	if m.Player.URL != videoURL {
		t.Fatalf("Player.URL = %q, want %q", m.Player.URL, videoURL)
	}
}

func TestModelInit_NoOptionsBaseBatchShape(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil cmd")
	}
	if m.Download.DownloadManager != m.DownloadManager {
		t.Fatalf("download manager not wired by Init()")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() cmd() type = %T, want tea.BatchMsg", msg)
	}
	if len(batch) != 4 {
		t.Fatalf("base batch command count = %d, want 4", len(batch))
	}
}

func TestModelInit_ChannelOptionSetsLoadingState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModelWithOptions(&models.CLIOptions{Channel: "xdagiz"})
	cmd := m.Init()

	if cmd == nil {
		t.Fatalf("Init() returned nil cmd")
	}
	if m.State != types.StateLoading {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateLoading)
	}
	if m.LoadingType != "channel" {
		t.Fatalf("m.LoadingType = %q, want channel", m.LoadingType)
	}
	if !m.VideoList.IsChannelSearch || m.VideoList.IsPlaylistSearch {
		t.Fatalf("channel flags not set correctly: channel=%v playlist=%v", m.VideoList.IsChannelSearch, m.VideoList.IsPlaylistSearch)
	}
	if m.VideoList.ChannelName != "xdagiz" {
		t.Fatalf("m.VideoList.ChannelName = %q, want xdagiz", m.VideoList.ChannelName)
	}
	if m.VideoList.PlaylistURL != "" {
		t.Fatalf("m.VideoList.PlaylistURL = %q, want empty", m.VideoList.PlaylistURL)
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() cmd() type = %T, want tea.BatchMsg", msg)
	}
	if len(batch) != 5 {
		t.Fatalf("batch command count = %d, want 5 when option cmd exists", len(batch))
	}
}

func TestModelInit_QueryOptionSetsLoadingAndCommand(t *testing.T) {
	setupAppTeaEnv(t)

	query := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	m := NewModelWithOptions(&models.CLIOptions{Query: query})
	cmd := m.Init()

	if m.State != types.StateLoading {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateLoading)
	}
	if m.LoadingType != "search" {
		t.Fatalf("m.LoadingType = %q, want search", m.LoadingType)
	}
	if m.CurrentQuery != query {
		t.Fatalf("m.CurrentQuery = %q, want %q", m.CurrentQuery, query)
	}
	if m.VideoList.IsChannelSearch || m.VideoList.IsPlaylistSearch {
		t.Fatalf("query should disable channel/playlist flags")
	}
	if m.VideoList.ChannelName != "" || m.VideoList.PlaylistName != "" || m.VideoList.PlaylistURL != "" {
		t.Fatalf("query path should clear channel/playlist metadata")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() cmd() type = %T, want tea.BatchMsg", msg)
	}
	if len(batch) != 5 {
		t.Fatalf("batch command count = %d, want 5", len(batch))
	}

	optionMsg := batch[4]()
	startFormat, ok := optionMsg.(types.StartFormatMsg)
	if !ok {
		t.Fatalf("option cmd msg type = %T, want types.StartFormatMsg for video query", optionMsg)
	}
	if startFormat.URL != query {
		t.Fatalf("StartFormatMsg.URL = %q, want %q", startFormat.URL, query)
	}
}

func TestModelInit_PlaylistOptionSetsLoadingState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModelWithOptions(&models.CLIOptions{Playlist: "PL123456789"})
	cmd := m.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil cmd")
	}

	if m.State != types.StateLoading {
		t.Fatalf("m.State = %q, want %q", m.State, types.StateLoading)
	}
	if m.LoadingType != "playlist" {
		t.Fatalf("m.LoadingType = %q, want playlist", m.LoadingType)
	}
	if !m.VideoList.IsPlaylistSearch || m.VideoList.IsChannelSearch {
		t.Fatalf("playlist flags not set correctly: playlist=%v channel=%v", m.VideoList.IsPlaylistSearch, m.VideoList.IsChannelSearch)
	}
	if m.CurrentQuery != "PL123456789" {
		t.Fatalf("m.CurrentQuery = %q, want PL123456789", m.CurrentQuery)
	}
	if m.VideoList.PlaylistName != "PL123456789" {
		t.Fatalf("m.VideoList.PlaylistName = %q, want PL123456789", m.VideoList.PlaylistName)
	}
	if m.VideoList.PlaylistURL != "https://www.youtube.com/playlist?list=PL123456789" {
		t.Fatalf("m.VideoList.PlaylistURL = %q, unexpected", m.VideoList.PlaylistURL)
	}
}

func TestModelInit_OptionPrecedenceQueryOverChannel(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModelWithOptions(&models.CLIOptions{
		Channel: "chan",
		Query:   "hello world",
	})
	_ = m.Init()

	if m.LoadingType != "search" {
		t.Fatalf("m.LoadingType = %q, want search (query should override channel)", m.LoadingType)
	}
	if m.VideoList.IsChannelSearch || m.VideoList.IsPlaylistSearch {
		t.Fatalf("query path should disable channel/playlist flags")
	}
	if m.VideoList.ChannelName != "" {
		t.Fatalf("m.VideoList.ChannelName = %q, want empty after query override", m.VideoList.ChannelName)
	}
}

func TestModelInit_OptionPrecedencePlaylistOverAll(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModelWithOptions(&models.CLIOptions{
		Channel:  "chan",
		Query:    "hello world",
		Playlist: "PL999",
	})
	_ = m.Init()

	if m.LoadingType != "playlist" {
		t.Fatalf("m.LoadingType = %q, want playlist (playlist should override other options)", m.LoadingType)
	}
	if !m.VideoList.IsPlaylistSearch || m.VideoList.IsChannelSearch {
		t.Fatalf("playlist flags not set correctly after precedence")
	}
	if m.CurrentQuery != "PL999" {
		t.Fatalf("m.CurrentQuery = %q, want PL999", m.CurrentQuery)
	}
}

func TestAppCancelDownloadAfterResumeClearsAllState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Test Video"}
	m.Download.Progress.SetPercent(50.0)
	m.Download.CurrentSpeed = "1.5 MB/s"
	m.Download.CurrentETA = "10:00"
	m.Download.Phase = "[download] 50.0%"
	m.Download.FileDestination = "/tmp/downloads/video.mp4"
	m.Download.FileExtension = "mp4"
	m.Download.Paused = true

	if !m.Download.Paused {
		t.Fatalf("Initial Download.Paused = false, want true")
	}

	updated, _ := m.Update(types.ResumeDownloadMsg{})
	m = updated.(*Model)

	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false after resume")
	}

	updated, _ = m.Update(types.CancelDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = false, want true after CancelDownloadMsg")
	}

	if m.State == types.StateDownload {
		t.Fatalf("m.State = %q, want different state after cancel", m.State)
	}
}

func TestAppCancelDownloadAfterResumeResetsProgress(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Test Video"}
	m.Download.Progress.SetPercent(75.0)
	m.Download.CurrentSpeed = "2.0 MB/s"
	m.Download.CurrentETA = "5:00"
	m.Download.Phase = "[download] 75.0%"
	m.Download.Paused = true

	updated, _ := m.Update(types.ResumeDownloadMsg{})
	m = updated.(*Model)

	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false after resume")
	}

	updated, _ = m.Update(types.CancelDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = false, want true after cancel")
	}
}

func TestAppCancelDownloadAfterResumeClearsDestination(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Test Video"}
	m.Download.Destination = "/tmp/downloads"
	m.Download.FileDestination = "/tmp/downloads/video.mp4"
	m.Download.FileExtension = "mp4"
	m.Download.Paused = true

	updated, _ := m.Update(types.ResumeDownloadMsg{})
	m = updated.(*Model)

	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false after resume")
	}

	updated, _ = m.Update(types.CancelDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = false, want true")
	}
}

func TestAppCancelDownloadAfterPauseResumeCycleClearsState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Test Video"}
	m.Download.Progress.SetPercent(25.0)
	m.Download.CurrentSpeed = "500 KB/s"
	m.Download.CurrentETA = "20:00"

	updated, _ := m.Update(types.PauseDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Paused {
		t.Fatalf("Download.Paused = false, want true after pause")
	}

	updated, _ = m.Update(types.ResumeDownloadMsg{})
	m = updated.(*Model)

	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false after resume")
	}

	updated, _ = m.Update(types.PauseDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Paused {
		t.Fatalf("Download.Paused = false, want true after second pause")
	}

	updated, _ = m.Update(types.ResumeDownloadMsg{})
	m = updated.(*Model)

	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false after second resume")
	}

	updated, _ = m.Update(types.CancelDownloadMsg{})
	m = updated.(*Model)

	if !m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = false, want true after cancel")
	}
}

func TestAppStartResumeDownloadClearsAllState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Old Video"}
	m.Download.Progress.SetPercent(75.0)
	m.Download.CurrentSpeed = "2.0 MB/s"
	m.Download.CurrentETA = "5:00"
	m.Download.Phase = "[download] 75.0% of 100%"
	m.Download.FileDestination = "/tmp/downloads/old-video.mp4"
	m.Download.FileExtension = "mp4"
	m.Download.Paused = true
	m.Download.Completed = false
	m.Download.Cancelled = false

	updated, _ := m.Update(types.StartResumeDownloadMsg{
		URL:      "https://youtube.com/watch?v=newvideo",
		URLs:     nil,
		Videos:   nil,
		FormatID: "best",
		Title:    "New Video",
	})
	m = updated.(*Model)

	if m.Download.CurrentSpeed != "" {
		t.Fatalf("Download.CurrentSpeed = %q, want empty", m.Download.CurrentSpeed)
	}
	if m.Download.CurrentETA != "" {
		t.Fatalf("Download.CurrentETA = %q, want empty", m.Download.CurrentETA)
	}
	if m.Download.Phase != "" {
		t.Fatalf("Download.Phase = %q, want empty", m.Download.Phase)
	}
	if m.Download.FileDestination != "" {
		t.Fatalf("Download.FileDestination = %q, want empty", m.Download.FileDestination)
	}
	if m.Download.FileExtension != "" {
		t.Fatalf("Download.FileExtension = %q, want empty", m.Download.FileExtension)
	}
	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false")
	}
	if m.Download.Completed {
		t.Fatalf("Download.Completed = true, want false")
	}
	if m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = true, want false")
	}
}

func TestAppStartDownloadClearsAllState(t *testing.T) {
	setupAppTeaEnv(t)

	m := NewModel()
	m.State = types.StateDownload
	m.Download.SelectedVideo = types.VideoItem{ID: "abc", VideoTitle: "Old Video"}
	m.Download.Progress.SetPercent(50.0)
	m.Download.CurrentSpeed = "1.0 MB/s"
	m.Download.CurrentETA = "10:00"
	m.Download.Phase = "[download] 50.0%"
	m.Download.FileDestination = "/tmp/downloads/old-video.mp4"
	m.Download.FileExtension = "mp4"
	m.Download.Paused = true
	m.Download.Completed = false
	m.Download.Cancelled = false

	updated, _ := m.Update(types.StartDownloadMsg{
		URL:           "https://youtube.com/watch?v=newvideo",
		FormatID:      "best",
		IsAudioTab:    false,
		ABR:           0,
		SelectedVideo: types.VideoItem{ID: "newvideo", VideoTitle: "New Video"},
	})
	m = updated.(*Model)

	if m.Download.CurrentSpeed != "" {
		t.Fatalf("Download.CurrentSpeed = %q, want empty", m.Download.CurrentSpeed)
	}
	if m.Download.CurrentETA != "" {
		t.Fatalf("Download.CurrentETA = %q, want empty", m.Download.CurrentETA)
	}
	if m.Download.Phase != "" {
		t.Fatalf("Download.Phase = %q, want empty", m.Download.Phase)
	}
	if m.Download.FileDestination != "" {
		t.Fatalf("Download.FileDestination = %q, want empty", m.Download.FileDestination)
	}
	if m.Download.FileExtension != "" {
		t.Fatalf("Download.FileExtension = %q, want empty", m.Download.FileExtension)
	}
	if m.Download.Paused {
		t.Fatalf("Download.Paused = true, want false")
	}
	if m.Download.Completed {
		t.Fatalf("Download.Completed = true, want false")
	}
	if m.Download.Cancelled {
		t.Fatalf("Download.Cancelled = true, want false")
	}
}
