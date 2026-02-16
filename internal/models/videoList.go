package models

import (
	"fmt"
	"log"
	"strings"

	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/styles"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type VideoListModel struct {
	Width            int
	Height           int
	List             list.Model
	CurrentQuery     string
	IsChannelSearch  bool
	IsPlaylistSearch bool
	ChannelName      string
	PlaylistName     string
	PlaylistURL      string
	ErrMsg           string
	DownloadOptions  []types.DownloadOption
	SelectedVideos   []types.VideoItem
}

func NewVideoListModel() VideoListModel {
	dl := styles.NewListDelegate()
	li := list.New([]list.Item{}, dl, 0, 0)
	li.SetShowStatusBar(false)
	li.SetShowTitle(false)
	li.SetShowHelp(false)
	li.KeyMap.Quit.SetKeys("q")
	li.FilterInput.Cursor.Style = li.FilterInput.Cursor.Style.Foreground(styles.MauveColor)
	li.FilterInput.PromptStyle = li.FilterInput.PromptStyle.Foreground(styles.SecondaryColor)

	return VideoListModel{
		List:             li,
		IsChannelSearch:  false,
		IsPlaylistSearch: false,
		ChannelName:      "",
		PlaylistName:     "",
		PlaylistURL:      "",
		ErrMsg:           "",
	}
}

func (m VideoListModel) Init() tea.Cmd {
	return nil
}

func (m VideoListModel) View() string {
	var (
		s           strings.Builder
		headerText  string
		headerStyle lipgloss.Style
	)

	if m.ErrMsg != "" {
		headerStyle = styles.ErrorMessageStyle.PaddingTop(1)
		if strings.Contains(m.ErrMsg, "Channel not found") {
			headerText = fmt.Sprintf("Channel not found: @%s", m.ChannelName)
		} else if strings.Contains(m.ErrMsg, "Playlist not found") {
			headerText = fmt.Sprintf("Playlist not found: %s", m.PlaylistName)
		} else if strings.Contains(m.ErrMsg, "private") {
			headerText = fmt.Sprintf("Private playlist: %s", m.PlaylistName)
		} else {
			headerText = fmt.Sprintf("An Error Occured: %s", m.ErrMsg)
		}
	} else if m.IsChannelSearch {
		headerText = fmt.Sprintf("Videos for channel @%s", m.ChannelName)
		headerStyle = styles.SectionHeaderStyle
	} else if m.IsPlaylistSearch {
		headerText = fmt.Sprintf("Playlist: %s", m.PlaylistName)
		headerStyle = styles.SectionHeaderStyle
	} else {
		headerText = fmt.Sprintf("Search Results for: %s", m.CurrentQuery)
		headerStyle = styles.SectionHeaderStyle
	}

	s.WriteString(headerStyle.Render(headerText))
	s.WriteRune('\n')
	s.WriteString(styles.ListContainer.Render(m.List.View()))

	return s.String()
}

func (m VideoListModel) HandleResize(w, h int) VideoListModel {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h-7)
	return m
}

func (m VideoListModel) isVideoSelected(video types.VideoItem) bool {
	for _, v := range m.SelectedVideos {
		if v.ID == video.ID {
			return true
		}
	}

	return false
}

func (m *VideoListModel) UpdateListItems() {
	items := m.List.Items()
	newItems := make([]list.Item, len(items))

	for i, item := range items {
		if video, ok := item.(types.SelectableVideoItem); ok {
			video.IsSelected = m.isVideoSelected(video.VideoItem)
			newItems[i] = video
		} else if video, ok := item.(types.VideoItem); ok {
			newItems[i] = types.SelectableVideoItem{
				VideoItem:  video,
				IsSelected: m.isVideoSelected(video),
			}
		} else {
			newItems[i] = item
		}
	}

	m.List.SetItems(newItems)
}

func (m VideoListModel) Update(msg tea.Msg) (VideoListModel, tea.Cmd) {
	var (
		cmd     tea.Cmd
		listCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			if !m.List.SettingFilter() {
				if m.ErrMsg != "" || len(m.List.Items()) == 0 {
					return m, nil
				}

				cfg, err := config.Load()
				if err != nil {
					log.Printf("Warning: Failed to load config: %v", err)
				}
				formatID := cfg.GetDefaultFormat()

				if len(m.SelectedVideos) > 0 {
					cmd = func() tea.Msg {
						return types.StartQueueDownloadMsg{
							Videos:          m.SelectedVideos,
							FormatID:        formatID,
							IsAudioTab:      false,
							ABR:             0,
							DownloadOptions: m.DownloadOptions,
						}
					}

					return m, cmd
				}

				selectedItem := m.List.SelectedItem()
				var video types.VideoItem
				if sv, ok := selectedItem.(types.SelectableVideoItem); ok {
					video = sv.VideoItem
				} else if v, ok := selectedItem.(types.VideoItem); ok {
					video = v
				} else {
					return m, nil
				}

				var url string
				if m.IsPlaylistSearch && m.PlaylistURL != "" {
					url = utils.BuildPlaylistURL(m.PlaylistURL)
				} else {
					url = utils.BuildVideoURL(video.ID)
				}

				cmd = func() tea.Msg {
					return types.StartDownloadMsg{
						URL:             url,
						FormatID:        formatID,
						SelectedVideo:   video,
						DownloadOptions: m.DownloadOptions,
					}
				}
			}
		}

		switch msg.Type {
		case tea.KeyEnter:
			if m.List.SettingFilter() {
				m.List.SetFilterState(list.FilterApplied)
				return m, nil
			}
			if m.ErrMsg != "" {
				cmd = func() tea.Msg {
					return types.BackFromVideoListMsg{}
				}
				return m, cmd
			} else if len(m.List.Items()) == 0 {
				return m, nil
			}

			selectedItem := m.List.SelectedItem()
			var video types.VideoItem
			if sv, ok := selectedItem.(types.SelectableVideoItem); ok {
				video = sv.VideoItem
			} else if v, ok := selectedItem.(types.VideoItem); ok {
				video = v
			} else {
				return m, nil
			}

			if video.ID == "" {
				return m, nil
			}

			if len(m.SelectedVideos) > 0 {
				cmd = func() tea.Msg {
					return types.StartQueueConfirmMsg{Videos: m.SelectedVideos}
				}
			} else {
				var url string
				if m.IsPlaylistSearch && m.PlaylistURL != "" {
					url = utils.BuildPlaylistURL(m.PlaylistURL)
				} else {
					url = utils.BuildVideoURL(video.ID)
				}

				cmd = func() tea.Msg {
					return types.StartFormatMsg{URL: url, SelectedVideo: video}
				}
			}

		case tea.KeySpace:
			if m.ErrMsg == "" {
				selectedItem := m.List.SelectedItem()
				var video types.VideoItem

				if sv, ok := selectedItem.(types.SelectableVideoItem); ok {
					video = sv.VideoItem
				} else if v, ok := selectedItem.(types.VideoItem); ok {
					video = v
				}

				m.SelectedVideos = toggleVideoSelection(m.SelectedVideos, video)
				m.UpdateListItems()
				log.Print("SelectedVideos: ", m.SelectedVideos)
			}
		}

		switch msg.String() {
		case "a":
			if m.ErrMsg == "" {
				m.SelectAll()
			}
		}
	}

	m.List, listCmd = m.List.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func ToggleVideoSelection(selected []types.VideoItem, video types.VideoItem) []types.VideoItem {
	return toggleVideoSelection(selected, video)
}

func toggleVideoSelection(selected []types.VideoItem, video types.VideoItem) []types.VideoItem {
	for i, v := range selected {
		if v.ID == video.ID {
			return append(selected[:i], selected[i+1:]...)
		}
	}

	return append(selected, video)
}

func (m VideoListModel) GetSelectedVideos() []types.VideoItem {
	return m.SelectedVideos
}

func (m *VideoListModel) ClearSelection() {
	m.SelectedVideos = nil
	m.UpdateListItems()
}

func (m VideoListModel) HasSelection() bool {
	return len(m.SelectedVideos) > 0
}

func (m *VideoListModel) SelectAll() {
	items := m.List.Items()

	allVideos := make([]types.VideoItem, 0, len(items))
	for _, item := range items {
		if sv, ok := item.(types.SelectableVideoItem); ok {
			allVideos = append(allVideos, sv.VideoItem)
		} else if v, ok := item.(types.VideoItem); ok {
			allVideos = append(allVideos, v)
		}
	}

	if len(m.SelectedVideos) == len(allVideos) && len(allVideos) > 0 {
		m.SelectedVideos = nil
	} else {
		m.SelectedVideos = allVideos
	}

	m.UpdateListItems()
}

func (m *VideoListModel) SetItems(items []list.Item) {
	selectableItems := make([]list.Item, len(items))
	for i, item := range items {
		if video, ok := item.(types.VideoItem); ok {
			selectableItems[i] = types.SelectableVideoItem{
				VideoItem:  video,
				IsSelected: false,
			}
		} else {
			selectableItems[i] = item
		}
	}

	m.List.SetItems(selectableItems)
}
