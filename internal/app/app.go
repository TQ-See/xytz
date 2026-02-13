package app

import (
	"github.com/xdagiz/xytz/internal/models"
	"github.com/xdagiz/xytz/internal/styles"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"
	"github.com/xdagiz/xytz/internal/version"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Program         *tea.Program
	Search          models.SearchModel
	State           types.State
	Width           int
	Height          int
	Spinner         spinner.Model
	LoadingType     string
	CurrentQuery    string
	Videos          []list.Item
	VideoList       models.VideoListModel
	FormatList      models.FormatListModel
	Download        models.DownloadModel
	SelectedVideo   types.VideoItem
	ErrMsg          string
	SearchManager   *utils.SearchManager
	FormatsManager  *utils.FormatsManager
	DownloadManager *utils.DownloadManager
	latestVersion   string
}

func (m *Model) Init() tea.Cmd {
	m.InitDownloadManager()
	opts := m.Search.Options
	var cmd tea.Cmd

	if opts != nil {
		if opts.Channel != "" {
			m.State = types.StateLoading
			m.LoadingType = "channel"
			m.VideoList.IsChannelSearch = true
			m.VideoList.IsPlaylistSearch = false
			m.VideoList.ChannelName = opts.Channel
			m.VideoList.PlaylistURL = ""
			cmd = utils.PerformChannelSearch(m.SearchManager, opts.Channel, m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		}

		if opts.Query != "" {
			m.State = types.StateLoading
			m.LoadingType = "search"
			m.CurrentQuery = opts.Query
			m.VideoList.IsChannelSearch = false
			m.VideoList.IsPlaylistSearch = false
			m.VideoList.ChannelName = ""
			m.VideoList.PlaylistName = ""
			m.VideoList.PlaylistURL = ""
			cmd = utils.PerformSearch(m.SearchManager, opts.Query, m.Search.SortBy.GetSPParam(), m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		}

		if opts.Playlist != "" {
			m.State = types.StateLoading
			m.LoadingType = "playlist"
			m.CurrentQuery = opts.Playlist
			m.VideoList.IsPlaylistSearch = true
			m.VideoList.IsChannelSearch = false
			m.VideoList.PlaylistName = opts.Playlist
			m.VideoList.PlaylistURL = utils.BuildPlaylistURL(opts.Playlist)
			cmd = utils.PerformPlaylistSearch(m.SearchManager, m.VideoList.PlaylistURL, m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		}
	}

	return tea.Batch(m.Search.Init(), m.Spinner.Tick, m.Download.Init(), m.fetchLatestVersion(), cmd)
}

func (m *Model) InitDownloadManager() {
	m.Download.DownloadManager = m.DownloadManager
}

func NewModel() *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = sp.Style.Foreground(styles.PinkColor)

	return &Model{
		State:           types.StateSearchInput,
		Spinner:         sp,
		Search:          models.NewSearchModel(),
		VideoList:       models.NewVideoListModel(),
		FormatList:      models.NewFormatListModel(),
		Download:        models.NewDownloadModel(),
		SearchManager:   utils.NewSearchManager(),
		FormatsManager:  utils.NewFormatsManager(),
		DownloadManager: utils.NewDownloadManager(),
	}
}

func NewModelWithOptions(opts *models.CLIOptions) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = sp.Style.Foreground(styles.PinkColor)

	return &Model{
		State:           types.StateSearchInput,
		Spinner:         sp,
		Search:          models.NewSearchModelWithOptions(opts),
		VideoList:       models.NewVideoListModel(),
		FormatList:      models.NewFormatListModel(),
		Download:        models.NewDownloadModel(),
		SearchManager:   utils.NewSearchManager(),
		FormatsManager:  utils.NewFormatsManager(),
		DownloadManager: utils.NewDownloadManager(),
	}
}

type latestVersionMsg struct {
	version string
	err     error
}

func (m *Model) fetchLatestVersion() tea.Cmd {
	return func() tea.Msg {
		version, err := version.FetchLatestVersion()
		return latestVersionMsg{version: version, err: err}
	}
}
