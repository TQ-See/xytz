package models

import (
	"strings"

	"github.com/xdagiz/xytz/internal/styles"
	"github.com/xdagiz/xytz/internal/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type FormatTab int

const (
	FormatTabVideo FormatTab = iota
	FormatTabAudio
	FormatTabThumbnail
	FormatTabCustom
)

var formatTabNames = []string{"Video", "Audio", "Thumbnail", "Custom"}

type FormatListModel struct {
	Width            int
	Height           int
	List             list.Model
	URL              string
	DownloadOptions  []types.DownloadOption
	ActiveTab        FormatTab
	VideoFormats     []list.Item
	AudioFormats     []list.Item
	ThumbnailFormats []list.Item
	AllFormats       []list.Item
}

func NewFormatListModel() FormatListModel {
	fd := list.NewDefaultDelegate()
	fd.Styles.NormalTitle = styles.ListTitleStyle
	fd.Styles.SelectedTitle = styles.ListSelectedTitleStyle
	fd.Styles.NormalDesc = styles.ListDescStyle
	fd.Styles.SelectedDesc = styles.ListSelectedDescStyle
	li := list.New([]list.Item{}, fd, 0, 0)
	li.SetShowStatusBar(false)
	li.SetShowTitle(false)
	li.FilterInput.Cursor.Style = li.FilterInput.Cursor.Style.Foreground(styles.PinkColor)
	li.FilterInput.PromptStyle = li.FilterInput.PromptStyle.Foreground(styles.SecondaryColor)

	return FormatListModel{
		List:      li,
		ActiveTab: FormatTabVideo,
	}
}

func (m FormatListModel) Init() tea.Cmd {
	return nil
}

func (m FormatListModel) View() string {
	var s strings.Builder

	s.WriteString(styles.SectionHeaderStyle.Foreground(styles.MauveColor).Padding(1, 0).Render("Select a Format"))
	s.WriteRune('\n')

	container := styles.FormatContainerStyle
	s.WriteString(container.Render(m.renderTabs()))
	s.WriteRune('\n')

	if m.ActiveTab == FormatTabCustom {
		s.WriteString(container.PaddingLeft(4).Render(styles.FormatCustomMessageStyle.Render("Custom format selection coming soon...")))
	} else {
		s.WriteString(container.Render(styles.ListContainer.Render(m.List.View())))
	}

	return s.String()
}

func (m FormatListModel) renderTabs() string {
	var tabBar strings.Builder

	for i, name := range formatTabNames {
		var style = styles.TabInactiveStyle
		if FormatTab(i) == m.ActiveTab {
			style = styles.TabActiveStyle
		}

		if i > 0 {
			tabBar.WriteString(" ")
		}

		tabBar.WriteString(style.Render(" " + name + " "))
	}

	tabBar.WriteString(styles.FormatTabHelpStyle.Render("   (tab to switch)"))

	return tabBar.String()
}

func (m FormatListModel) HandleResize(w, h int) FormatListModel {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h-8)
	return m
}

func (m FormatListModel) Update(msg tea.Msg) (FormatListModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, formatTabNext):
			m.nextTab()
			return m, nil
		case key.Matches(msg, formatTabPrev):
			m.prevTab()
			return m, nil
		case msg.Type == tea.KeyEnter:
			if m.ActiveTab == FormatTabCustom {
				return m, nil
			}
			if m.List.FilterState() == list.Filtering {
				m.List.SetFilterState(list.FilterApplied)
				return m, nil
			}
			if len(m.List.Items()) == 0 {
				return m, nil
			}
			item := m.List.SelectedItem()
			if item == nil {
				return m, nil
			}
			format, ok := item.(types.FormatItem)
			if !ok {
				return m, nil
			}
			cmd = func() tea.Msg {
				msg := types.StartDownloadMsg{
					URL:             m.URL,
					FormatID:        format.FormatValue,
					DownloadOptions: m.DownloadOptions,
				}
				return msg
			}
		}
	}

	var listCmd tea.Cmd
	m.List, listCmd = m.List.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m *FormatListModel) nextTab() {
	m.ActiveTab++
	if m.ActiveTab > FormatTabCustom {
		m.ActiveTab = FormatTabVideo
	}

	m.updateListForTab()
}

func (m *FormatListModel) prevTab() {
	m.ActiveTab--
	if m.ActiveTab < FormatTabVideo {
		m.ActiveTab = FormatTabCustom
	}

	m.updateListForTab()
}

func (m *FormatListModel) updateListForTab() {
	switch m.ActiveTab {
	case FormatTabVideo:
		m.List.SetItems(m.VideoFormats)
	case FormatTabAudio:
		m.List.SetItems(m.AudioFormats)
	case FormatTabThumbnail:
		m.List.SetItems(m.ThumbnailFormats)
	case FormatTabCustom:
		m.List.SetItems([]list.Item{})
	}

	m.List.ResetSelected()
}

func (m *FormatListModel) SetFormats(videoFormats, audioFormats, thumbnailFormats, allFormats []list.Item) {
	m.VideoFormats = videoFormats
	m.AudioFormats = audioFormats
	m.ThumbnailFormats = thumbnailFormats
	m.AllFormats = allFormats
	m.updateListForTab()
}

func (m *FormatListModel) ClearSelection() {
	m.List.Select(-1)
}

func (m *FormatListModel) ResetTab() {
	m.ActiveTab = FormatTabVideo
	m.updateListForTab()
}

var formatTabNext = key.NewBinding(key.WithKeys("tab"))
var formatTabPrev = key.NewBinding(key.WithKeys("shift+tab"))
