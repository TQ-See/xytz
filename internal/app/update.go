package app

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xdagiz/xytz/internal/models"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	queueLabel := strings.TrimSpace(m.Download.QueueLabel)
	if queueLabel == "" {
		queueLabel = strings.TrimSpace(m.CurrentQuery)
	}
	if queueLabel == "" {
		queueLabel = strings.TrimSpace(m.VideoList.PlaylistName)
	}

	if queueLabel == "" {
		queueLabel = "Queued downloads"
	}

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
		m.FormatList.IsQueue = false
		m.FormatList.QueueVideos = nil
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
		m.VideoList.SetItems(msg.Videos)
		m.VideoList.CurrentQuery = m.CurrentQuery
		m.VideoList.ErrMsg = msg.Err
		m.State = types.StateVideoList
		m.ErrMsg = msg.Err
		return m, nil

	case types.FormatResultMsg:
		m.LoadingType = ""
		m.FormatList.SetFormats(msg.VideoFormats, msg.AudioFormats, msg.ThumbnailFormats, msg.AllFormats)
		m.FormatList.ShowVideoInfo = !m.FormatList.IsQueue
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
		m.Download.QueueLabel = ""
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
			Title:              m.Download.SelectedVideo.Title(),
			Videos:             []types.VideoItem{m.Download.SelectedVideo},
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
		m.LoadingType = "download"
		resumeURLs := msg.URLs
		resumeVideos := msg.Videos
		resumeTitle := msg.Title
		resumeFormatID := msg.FormatID
		if len(resumeURLs) > 0 {
			videos := resumeVideos
			if len(videos) == 0 {
				videos = make([]types.VideoItem, len(resumeURLs))
				for i, u := range resumeURLs {
					videos[i] = types.VideoItem{ID: u, VideoTitle: u}
				}
			}

			queueLabel := resumeTitle
			if queueLabel == "" {
				queueLabel = "Queued downloads"
			}

			m.Download.IsQueue = true
			m.Download.QueueLabel = queueLabel
			m.Download.QueueTotal = len(videos)
			m.Download.QueueIndex = 1
			m.Download.SelectedVideo = videos[0]
			m.Download.QueueItems = make([]types.QueueItem, len(videos))
			m.Download.QueueFormatID = resumeFormatID
			m.Download.QueueIsAudioTab = false
			m.Download.QueueABR = 0

			for i, v := range videos {
				m.Download.QueueItems[i] = types.QueueItem{
					Index:  i + 1,
					Video:  v,
					URL:    v.ID,
					Status: types.QueueStatusPending,
				}
			}

			updateQueueUnfinished(queueLabel, resumeFormatID, m.Download.QueueTotal, pendingQueueURLs(m.Download.QueueItems), pendingQueueVideos(m.Download.QueueItems))

			m.Download.QueueItems[0].Status = types.QueueStatusDownloading
			req := types.DownloadRequest{
				URL:                m.Download.QueueItems[0].URL,
				URLs:               pendingQueueURLs(m.Download.QueueItems),
				Videos:             pendingQueueVideos(m.Download.QueueItems),
				FormatID:           resumeFormatID,
				IsAudioTab:         false,
				ABR:                0,
				QueueIndex:         1,
				QueueTotal:         m.Download.QueueTotal,
				UnfinishedKey:      utils.QueueUnfinishedKey(queueLabel),
				UnfinishedTitle:    queueLabel,
				UnfinishedDesc:     fmt.Sprintf("%d items left", m.Download.QueueTotal),
				Title:              m.Download.SelectedVideo.Title(),
				Options:            m.Search.DownloadOptions,
				CookiesFromBrowser: m.Search.CookiesFromBrowser,
				Cookies:            m.Search.Cookies,
			}

			cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
			return m, cmd
		}

		m.Download.SelectedVideo = types.VideoItem{VideoTitle: resumeTitle}
		if len(resumeVideos) > 0 {
			m.Download.SelectedVideo = resumeVideos[0]
		} else if resumeTitle != "" {
			m.Download.SelectedVideo = types.VideoItem{
				ID:         msg.URL,
				VideoTitle: resumeTitle,
			}
		}
		req := types.DownloadRequest{
			URL:                msg.URL,
			FormatID:           resumeFormatID,
			IsAudioTab:         false,
			ABR:                0,
			Title:              m.Download.SelectedVideo.Title(),
			Videos:             []types.VideoItem{m.Download.SelectedVideo},
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}
		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

	case types.DownloadResultMsg:
		m.LoadingType = ""
		if m.Download.IsQueue {
			if len(m.Download.QueueItems) >= m.Download.QueueIndex {
				item := &m.Download.QueueItems[m.Download.QueueIndex-1]
				if msg.Destination != "" {
					item.Destination = msg.Destination
				}

				if msg.Err != "" {
					item.Status = types.QueueStatusError
					item.Error = msg.Err
				} else {
					item.Status = types.QueueStatusComplete
				}
			}

			if m.Download.QueueIndex < m.Download.QueueTotal {
				m.Download.QueueIndex++
				next := &m.Download.QueueItems[m.Download.QueueIndex-1]
				next.Status = types.QueueStatusDownloading
				m.Download.SelectedVideo = next.Video
				m.Download.Progress.SetPercent(0)
				m.Download.CurrentSpeed = ""
				m.Download.CurrentETA = ""
				m.Download.Phase = ""

				remaining := queueRemaining(m.Download.QueueItems)
				updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, remaining, pendingQueueURLs(m.Download.QueueItems), pendingQueueVideos(m.Download.QueueItems))

				req := types.DownloadRequest{
					URL:                next.URL,
					FormatID:           m.Download.QueueFormatID,
					IsAudioTab:         m.Download.QueueIsAudioTab,
					ABR:                m.Download.QueueABR,
					QueueIndex:         m.Download.QueueIndex,
					QueueTotal:         m.Download.QueueTotal,
					URLs:               pendingQueueURLs(m.Download.QueueItems),
					Videos:             pendingQueueVideos(m.Download.QueueItems),
					UnfinishedKey:      utils.QueueUnfinishedKey(m.Download.QueueLabel),
					UnfinishedTitle:    m.Download.QueueLabel,
					UnfinishedDesc:     fmt.Sprintf("%d items left", remaining),
					Title:              next.Video.Title(),
					Options:            m.Search.DownloadOptions,
					CookiesFromBrowser: m.Search.CookiesFromBrowser,
					Cookies:            m.Search.Cookies,
				}

				cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
				return m, cmd
			}

			updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, 0, nil, nil)
			m.Download.QueueError = msg.Err
			m.Download.Completed = true

			return m, nil
		}

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
		if m.Download.IsQueue {
			urls := pendingQueueURLs(m.Download.QueueItems)
			videos := pendingQueueVideos(m.Download.QueueItems)
			remaining := queueRemaining(m.Download.QueueItems)
			if remaining == 0 && len(urls) > 0 {
				remaining = len(urls)
			}
			if len(urls) == 0 {
				updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, 0, nil, nil)
			} else {
				updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, remaining, urls, videos)
			}
		}

		m.State = types.StateSearchInput
		m.Search.Input.SetValue("")
		m.clearSelections()
		m.resetDownloadState()
		return m, nil

	case types.PauseDownloadMsg:
		m.Download.Paused = true
		return m, nil

	case types.ResumeDownloadMsg:
		m.Download.Paused = false
		return m, nil

	case types.CancelDownloadMsg:
		m.Download.Cancelled = true
		if m.Download.IsQueue {
			for i := m.Download.QueueIndex - 1; i < len(m.Download.QueueItems); i++ {
				if m.Download.QueueItems[i].Status == types.QueueStatusDownloading {
					m.Download.QueueItems[i].Status = types.QueueStatusPending
				}
			}

			remaining := queueRemaining(m.Download.QueueItems)
			urls := pendingQueueURLs(m.Download.QueueItems)
			if remaining == 0 && len(urls) > 0 {
				remaining = len(urls)
			}

			updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, remaining, urls, pendingQueueVideos(m.Download.QueueItems))
			m.Download.Completed = true
			return m, nil
		}

		if m.SelectedVideo.ID == "" {
			m.State = types.StateSearchInput
		} else {
			m.State = types.StateVideoList
		}

		m.ErrMsg = "Download cancelled"
		m.FormatList.List.ResetSelected()
		return m, nil

	case types.SkipCurrentQueueItemMsg:
		if !m.Download.IsQueue {
			return m, nil
		}

		m.Download.QueueItems[m.Download.QueueIndex-1].Status = types.QueueStatusSkipped
		m.Download.QueueError = ""

		if m.Download.QueueIndex < m.Download.QueueTotal {
			m.Download.QueueIndex++
			m.Download.QueueItems[m.Download.QueueIndex-1].Status = types.QueueStatusDownloading
			m.Download.SelectedVideo = m.Download.QueueItems[m.Download.QueueIndex-1].Video
			m.Download.Progress.SetPercent(0)
			m.Download.CurrentSpeed = ""
			m.Download.CurrentETA = ""
			m.Download.Phase = ""

			remaining := queueRemaining(m.Download.QueueItems)
			updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, remaining, pendingQueueURLs(m.Download.QueueItems), pendingQueueVideos(m.Download.QueueItems))

			req := types.DownloadRequest{
				URL:                m.Download.QueueItems[m.Download.QueueIndex-1].URL,
				URLs:               pendingQueueURLs(m.Download.QueueItems),
				Videos:             pendingQueueVideos(m.Download.QueueItems),
				FormatID:           m.Download.QueueFormatID,
				IsAudioTab:         m.Download.QueueIsAudioTab,
				ABR:                m.Download.QueueABR,
				QueueIndex:         m.Download.QueueIndex,
				QueueTotal:         m.Download.QueueTotal,
				UnfinishedKey:      utils.QueueUnfinishedKey(m.Download.QueueLabel),
				UnfinishedTitle:    m.Download.QueueLabel,
				UnfinishedDesc:     fmt.Sprintf("%d items left", remaining),
				Title:              m.Download.QueueItems[m.Download.QueueIndex-1].Video.Title(),
				Options:            m.Search.DownloadOptions,
				CookiesFromBrowser: m.Search.CookiesFromBrowser,
				Cookies:            m.Search.Cookies,
			}

			cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
			return m, cmd
		}

		updateQueueUnfinished(queueLabel, m.Download.QueueFormatID, 0, nil, nil)
		m.Download.Completed = true
		return m, nil

	case types.RetryCurrentQueueItemMsg:
		if !m.Download.IsQueue {
			return m, nil
		}

		m.Download.QueueItems[m.Download.QueueIndex-1].Status = types.QueueStatusDownloading
		m.Download.QueueItems[m.Download.QueueIndex-1].Error = ""
		m.Download.QueueError = ""
		m.Download.Progress.SetPercent(0)
		m.Download.CurrentSpeed = ""
		m.Download.CurrentETA = ""
		m.Download.Phase = ""

		remaining := queueRemaining(m.Download.QueueItems)

		req := types.DownloadRequest{
			URL:                m.Download.QueueItems[m.Download.QueueIndex-1].URL,
			URLs:               pendingQueueURLs(m.Download.QueueItems),
			Videos:             pendingQueueVideos(m.Download.QueueItems),
			FormatID:           m.Download.QueueFormatID,
			IsAudioTab:         m.Download.QueueIsAudioTab,
			ABR:                m.Download.QueueABR,
			QueueIndex:         m.Download.QueueIndex,
			QueueTotal:         m.Download.QueueTotal,
			UnfinishedKey:      utils.QueueUnfinishedKey(m.Download.QueueLabel),
			UnfinishedTitle:    m.Download.QueueLabel,
			UnfinishedDesc:     fmt.Sprintf("%d items left", remaining),
			Title:              m.Download.QueueItems[m.Download.QueueIndex-1].Video.Title(),
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}

		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

	case types.CancelSearchMsg:
		m.State = types.StateSearchInput
		m.LoadingType = ""
		m.ErrMsg = "Search cancelled"
		m.clearSelections()
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
		m.clearSelections()
		m.VideoList.ErrMsg = ""
		m.VideoList.PlaylistURL = ""
		return m, nil

	case types.StartQueueConfirmMsg:
		if m.DownloadManager != nil {
			_ = m.DownloadManager.Cancel()
		}
		m.resetDownloadState()
		m.State = types.StateLoading
		m.LoadingType = "format"
		m.FormatList.IsQueue = true
		m.FormatList.QueueVideos = msg.Videos
		m.FormatList.DownloadOptions = m.Search.DownloadOptions
		m.FormatList.ShowVideoInfo = false
		first := msg.Videos[0]
		m.FormatList.URL = utils.BuildVideoURL(first.ID)
		m.FormatList.SelectedVideo = first
		return m, utils.FetchFormats(m.FormatsManager, m.FormatList.URL)

	case types.StartQueueConfirmWithFormatMsg:
		if m.DownloadManager != nil {
			_ = m.DownloadManager.Cancel()
		}
		m.resetDownloadState()
		m.State = types.StateDownload
		m.LoadingType = "queue"
		m.Download.IsQueue = true
		m.Download.QueueLabel = queueLabel
		m.Download.QueueTotal = len(msg.Videos)
		m.Download.QueueIndex = 1
		m.Download.SelectedVideo = msg.Videos[0]
		m.Download.QueueItems = make([]types.QueueItem, len(msg.Videos))
		m.Download.QueueFormatID = msg.FormatID
		m.Download.QueueIsAudioTab = msg.IsAudioTab
		m.Download.QueueABR = msg.ABR

		for i, v := range msg.Videos {
			url := utils.BuildVideoURL(v.ID)
			m.Download.QueueItems[i] = types.QueueItem{
				Index:  i + 1,
				Video:  v,
				URL:    url,
				Status: types.QueueStatusPending,
			}
		}

		updateQueueUnfinished(queueLabel, msg.FormatID, m.Download.QueueTotal, pendingQueueURLs(m.Download.QueueItems), pendingQueueVideos(m.Download.QueueItems))

		m.Download.QueueItems[0].Status = types.QueueStatusDownloading

		req := types.DownloadRequest{
			URL:                m.Download.QueueItems[0].URL,
			URLs:               pendingQueueURLs(m.Download.QueueItems),
			Videos:             pendingQueueVideos(m.Download.QueueItems),
			FormatID:           msg.FormatID,
			IsAudioTab:         msg.IsAudioTab,
			ABR:                msg.ABR,
			QueueIndex:         1,
			QueueTotal:         m.Download.QueueTotal,
			UnfinishedKey:      utils.QueueUnfinishedKey(queueLabel),
			UnfinishedTitle:    queueLabel,
			UnfinishedDesc:     fmt.Sprintf("%d items left", m.Download.QueueTotal),
			Title:              m.Download.SelectedVideo.Title(),
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}

		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

	case types.StartQueueDownloadMsg:
		if m.DownloadManager != nil {
			_ = m.DownloadManager.Cancel()
		}
		m.resetDownloadState()
		m.State = types.StateDownload
		m.LoadingType = "queue"
		m.Download.IsQueue = true
		m.Download.QueueLabel = queueLabel
		sourceVideos := msg.Videos
		m.Download.QueueTotal = len(sourceVideos)
		m.Download.QueueIndex = 1
		m.Download.SelectedVideo = sourceVideos[0]
		m.Download.QueueItems = make([]types.QueueItem, len(sourceVideos))
		m.Download.QueueFormatID = msg.FormatID
		m.Download.QueueIsAudioTab = msg.IsAudioTab
		m.Download.QueueABR = msg.ABR

		for i, v := range sourceVideos {
			url := utils.BuildVideoURL(v.ID)
			m.Download.QueueItems[i] = types.QueueItem{
				Index:  i + 1,
				Video:  v,
				URL:    url,
				Status: types.QueueStatusPending,
			}
		}

		updateQueueUnfinished(queueLabel, msg.FormatID, m.Download.QueueTotal, pendingQueueURLs(m.Download.QueueItems), pendingQueueVideos(m.Download.QueueItems))

		m.Download.QueueItems[0].Status = types.QueueStatusDownloading

		req := types.DownloadRequest{
			URL:                m.Download.QueueItems[0].URL,
			URLs:               pendingQueueURLs(m.Download.QueueItems),
			Videos:             pendingQueueVideos(m.Download.QueueItems),
			FormatID:           msg.FormatID,
			IsAudioTab:         msg.IsAudioTab,
			ABR:                msg.ABR,
			QueueIndex:         1,
			QueueTotal:         m.Download.QueueTotal,
			UnfinishedKey:      utils.QueueUnfinishedKey(queueLabel),
			UnfinishedTitle:    queueLabel,
			UnfinishedDesc:     fmt.Sprintf("%d items left", m.Download.QueueTotal),
			Title:              m.Download.SelectedVideo.Title(),
			Options:            m.Search.DownloadOptions,
			CookiesFromBrowser: m.Search.CookiesFromBrowser,
			Cookies:            m.Search.Cookies,
		}

		cmd = utils.StartDownload(m.DownloadManager, m.Program, req)
		return m, cmd

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
				if len(m.VideoList.SelectedVideos) > 0 {
					m.VideoList.ClearSelection()
					return m, nil
				} else {
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

			case " ":
				if m.VideoList.ErrMsg == "" {
					selectedItem := m.VideoList.List.SelectedItem()
					var video types.VideoItem

					if sv, ok := selectedItem.(types.SelectableVideoItem); ok {
						video = sv.VideoItem
					} else if v, ok := selectedItem.(types.VideoItem); ok {
						video = v
					}

					if video.ID != "" {
						m.VideoList.SelectedVideos = models.ToggleVideoSelection(m.VideoList.SelectedVideos, video)
						m.VideoList.UpdateListItems()
					}
				}

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
							m.clearSelections()
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
					m.clearSelections()
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

func queueRemaining(items []types.QueueItem) int {
	count := 0
	for _, it := range items {
		if it.Status == types.QueueStatusPending || it.Status == types.QueueStatusDownloading {
			count++
		}
	}

	return count
}

func pendingQueueURLs(items []types.QueueItem) []string {
	var urls []string
	for _, it := range items {
		if it.Status == types.QueueStatusPending || it.Status == types.QueueStatusDownloading || it.Status == types.QueueStatusError {
			if it.URL != "" {
				urls = append(urls, it.URL)
			}
		}
	}

	return urls
}

func pendingQueueVideos(items []types.QueueItem) []types.VideoItem {
	var videos []types.VideoItem
	for _, it := range items {
		if it.Status == types.QueueStatusPending || it.Status == types.QueueStatusDownloading || it.Status == types.QueueStatusError {
			if it.Video.ID != "" || it.Video.VideoTitle != "" {
				videos = append(videos, it.Video)
			}
		}
	}

	return videos
}

func (m *Model) clearSelections() {
	m.SelectedVideo = types.VideoItem{}
	m.VideoList.ClearSelection()
	m.VideoList.List.ResetSelected()
}

func updateQueueUnfinished(query, formatID string, remaining int, urls []string, videos []types.VideoItem) {
	label := strings.TrimSpace(query)
	if label == "" {
		label = "Queued downloads"
	}

	key := utils.QueueUnfinishedKey(label)
	if remaining <= 0 {
		if err := utils.RemoveUnfinished(key); err != nil {
			log.Printf("Failed to remove unfinished queue entry: %v", err)
		}

		return
	}

	if len(urls) == 0 {
		return
	}

	desc := fmt.Sprintf("%d items left", remaining)
	entry := utils.UnfinishedDownload{
		URL:       key,
		FormatID:  formatID,
		Title:     label,
		Desc:      desc,
		URLs:      urls,
		Videos:    videos,
		Timestamp: time.Now(),
	}

	if err := utils.AddUnfinished(entry); err != nil {
		log.Printf("Failed to update unfinished queue entry: %v", err)
	}
}

func (m *Model) resetDownloadState() {
	m.Download = models.NewDownloadModel()
	m.InitDownloadManager()
	m.SelectedVideo = types.VideoItem{}
	m.Download.QueueError = ""
	m.Download.IsQueue = false
	m.Download.QueueItems = nil
	m.Download.QueueIndex = 0
	m.Download.QueueTotal = 0
	m.Download.QueueFormatID = ""
	m.Download.QueueLabel = ""
	m.Download.QueueIsAudioTab = false
	m.Download.QueueABR = 0
	m.Download.QueueItems = nil
	m.Download.Progress.SetPercent(0)
	m.Download.CurrentSpeed = ""
	m.Download.CurrentETA = ""
	m.Download.Phase = ""
	m.Download.Completed = false
	m.Download.Cancelled = false
	m.Download.Paused = false
	m.FormatList.IsQueue = false
	m.FormatList.QueueVideos = nil
}
