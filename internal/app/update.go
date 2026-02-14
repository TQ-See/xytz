package app

import (
	"strings"

	"github.com/xdagiz/xytz/internal/models"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Search = m.Search.HandleResize(m.Width, m.Height)
		m.VideoList = m.VideoList.HandleResize(m.Width, m.Height)
		m.FormatList = m.FormatList.HandleResize(m.Width, m.Height)
		m.Download = m.Download.HandleResize(m.Width, m.Height)

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		return m, spinnerCmd

	case latestVersionMsg:
		if msg.err == nil {
			m.latestVersion = msg.version
			m.Search.LatestVersion = msg.version
		}

	case types.StartSearchMsg:
		m.State = types.StateLoading
		urlType, _ := utils.ParseSearchQuery(msg.Query)
		m.LoadingType = urlType
		m.CurrentQuery = strings.TrimSpace(msg.Query)
		m.VideoList.IsChannelSearch = urlType == "channel"
		m.VideoList.IsPlaylistSearch = urlType == "playlist"
		if urlType == "channel" {
			m.VideoList.ChannelName = utils.ExtractChannelUsername(msg.Query)
		}
		m.VideoList.PlaylistName = ""
		m.VideoList.PlaylistURL = ""
		cmd = utils.PerformSearch(m.SearchManager, msg.Query, m.Search.SortBy.GetSPParam(), m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		m.ErrMsg = ""
		m.Search.ErrMsg = ""
		m.Search.Input.SetValue("")

	case types.StartFormatMsg:
		m.State = types.StateLoading
		m.LoadingType = "format"
		m.FormatList.URL = msg.URL
		m.FormatList.SelectedVideo = msg.SelectedVideo
		m.SelectedVideo = msg.SelectedVideo
		m.FormatList.DownloadOptions = m.Search.DownloadOptions
		m.FormatList.ResetTab()
		cmd = utils.FetchFormats(m.FormatsManager, msg.URL)
		m.ErrMsg = ""

	case types.SearchResultMsg:
		m.LoadingType = ""
		m.Videos = msg.Videos
		m.VideoList.List.SetItems(msg.Videos)
		m.VideoList.CurrentQuery = m.CurrentQuery
		m.VideoList.ErrMsg = msg.Err
		m.State = types.StateVideoList
		m.ErrMsg = msg.Err
		return m, nil

	case types.FormatResultMsg:
		m.LoadingType = ""
		m.FormatList.SetFormats(msg.VideoFormats, msg.AudioFormats, msg.ThumbnailFormats, msg.AllFormats)
		if msg.VideoInfo.ID != "" {
			m.FormatList.SelectedVideo = msg.VideoInfo
		}
		m.State = types.StateFormatList
		m.ErrMsg = msg.Err
		return m, nil

	case types.StartDownloadMsg:
		m.State = types.StateDownload
		m.Download.Completed = false
		m.Download.Cancelled = false
		if msg.SelectedVideo.ID != "" {
			m.Download.SelectedVideo = msg.SelectedVideo
		} else if m.SelectedVideo.ID == "" {
			m.Download.SelectedVideo = m.FormatList.SelectedVideo
		} else {
			m.Download.SelectedVideo = m.SelectedVideo
		}
		m.LoadingType = "download"
		req := types.DownloadRequest{
			URL:                msg.URL,
			FormatID:           msg.FormatID,
			IsAudioTab:         msg.IsAudioTab,
			ABR:                msg.ABR,
			Title:              msg.SelectedVideo.Title(),
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}
		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

	case types.StartResumeDownloadMsg:
		m.State = types.StateDownload
		m.Download.Completed = false
		m.Download.Cancelled = false
		m.Download.SelectedVideo = types.VideoItem{VideoTitle: msg.Title}
		m.LoadingType = "download"
		req := types.DownloadRequest{
			URL:                msg.URL,
			FormatID:           msg.FormatID,
			IsAudioTab:         false,
			ABR:                0,
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}
		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

	case types.DownloadResultMsg:
		m.LoadingType = ""
		if msg.Err != "" {
			if !m.Download.Cancelled {
				m.ErrMsg = msg.Err
				m.State = types.StateSearchInput
			}
		} else {
			m.Download.Completed = true
		}
		return m, nil

	case types.DownloadCompleteMsg:
		m.State = types.StateSearchInput
		m.Search.Input.SetValue("")
		m.SelectedVideo = types.VideoItem{}
		m.Download.Progress.SetPercent(0)
		m.Download.CurrentSpeed = ""
		m.Download.CurrentETA = ""
		return m, nil

	case types.PauseDownloadMsg:
		m.Download.Paused = true
		return m, nil

	case types.ResumeDownloadMsg:
		m.Download.Paused = false
		return m, nil

	case types.CancelDownloadMsg:
		m.Download.Cancelled = true
		if m.SelectedVideo.ID == "" {
			m.State = types.StateSearchInput
		} else {
			m.State = types.StateVideoList
		}
		m.ErrMsg = "Download cancelled"
		m.FormatList.List.ResetSelected()
		return m, nil

	case types.CancelSearchMsg:
		m.State = types.StateSearchInput
		m.LoadingType = ""
		m.ErrMsg = "Search cancelled"
		return m, nil

	case types.CancelFormatsMsg:
		m.State = types.StateVideoList
		m.LoadingType = ""
		m.ErrMsg = ""
		m.FormatList.List.ResetSelected()
		return m, nil

	case types.StartChannelURLMsg:
		m.State = types.StateLoading
		m.LoadingType = "channel"
		m.VideoList.IsChannelSearch = true
		m.VideoList.IsPlaylistSearch = false
		m.VideoList.ChannelName = msg.ChannelName
		m.VideoList.PlaylistURL = ""
		cmd = utils.PerformChannelSearch(m.SearchManager, msg.ChannelName, m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		m.ErrMsg = ""
		return m, cmd

	case types.StartPlaylistURLMsg:
		m.State = types.StateLoading
		m.LoadingType = "playlist"
		m.CurrentQuery = strings.TrimSpace(msg.Query)
		m.VideoList.IsPlaylistSearch = true
		m.VideoList.IsChannelSearch = false
		m.VideoList.PlaylistName = strings.TrimSpace(msg.Query)
		m.VideoList.PlaylistURL = utils.BuildPlaylistURL(msg.Query)
		cmd = utils.PerformPlaylistSearch(m.SearchManager, msg.Query, m.Search.SearchLimit, m.Search.CookiesFromBrowser, m.Search.Cookies)
		m.ErrMsg = ""
		return m, cmd

	case types.BackFromVideoListMsg:
		m.State = types.StateSearchInput
		m.ErrMsg = ""
		m.SelectedVideo = types.VideoItem{}
		m.VideoList.List.ResetSelected()
		m.VideoList.ErrMsg = ""
		m.VideoList.PlaylistURL = ""
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		switch m.State {
		case types.StateSearchInput:
			m.Search, cmd = m.Search.Update(msg)
			m.ErrMsg = ""

		case types.StateLoading:
			switch msg.String() {
			case "c", "esc":
				switch m.LoadingType {
				case "format":
					cmd = utils.CancelFormats(m.FormatsManager)
				default:
					cmd = utils.CancelSearch(m.SearchManager)
				}
			}

		case types.StateVideoList:
			switch msg.String() {
			case "b", "esc":
				if HandleListEsc(m.VideoList.List) {
					m.State = types.StateSearchInput
					m.ErrMsg = ""
					m.VideoList.List.ResetFilter()
					m.VideoList.List.Select(0)
					return m, nil
				}

				m.VideoList.List.FilterInput.SetValue("")
				m.VideoList.List.SetFilterState(list.Unfiltered)
				return m, nil
			}
			m.VideoList, cmd = m.VideoList.Update(msg)

		case types.StateFormatList:
			switch msg.String() {
			case "b", "esc":
				if m.FormatList.ActiveTab != models.FormatTabCustom {
					if HandleListEsc(m.FormatList.List) {
						if m.SelectedVideo.ID == "" {
							m.State = types.StateSearchInput
							m.Search.Input.SetValue("")
						} else {
							m.State = types.StateVideoList
						}
						m.ErrMsg = ""
						m.FormatList.List.ResetFilter()
						m.FormatList.List.ResetSelected()
						return m, nil
					}

					m.VideoList.List.FilterInput.SetValue("")
					m.FormatList.List.SetFilterState(list.Unfiltered)
					return m, nil
				}
			}
			m.FormatList, cmd = m.FormatList.Update(msg)

		case types.StateDownload:
			switch msg.String() {
			case "b":
				if m.Download.Completed || m.Download.Cancelled {
					m.State = types.StateFormatList
					m.FormatList.List.ResetSelected()
				}

				m.ErrMsg = ""
				return m, nil
			}
		}

	case tea.MouseMsg:
		switch m.State {
		case types.StateSearchInput:
			m.Search, cmd = m.Search.Update(msg)
		}

	case list.FilterMatchesMsg:
		switch m.State {
		case types.StateSearchInput:
			m.Search, cmd = m.Search.Update(msg)
		case types.StateVideoList:
			m.VideoList, cmd = m.VideoList.Update(msg)
		case types.StateFormatList:
			m.FormatList, cmd = m.FormatList.Update(msg)
		}

		return m, cmd
	}

	switch m.State {
	case types.StateDownload:
		m.Download, cmd = m.Download.Update(msg)
	}

	return m, cmd
}

func HandleListEsc(l list.Model) bool {
	return models.HandleListEsc(l)
}
