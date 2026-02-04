package models

import (
	"strings"

	"github.com/xdagiz/xytz/internal/styles"
	"github.com/xdagiz/xytz/internal/utils"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

type ResumeKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Delete key.Binding
}

func DefaultResumeKeyMap() ResumeKeyMap {
	return ResumeKeyMap{
		Up: key.NewBinding(
			key.WithKeys("ctrl+p", "up"),
			key.WithHelp("Ctrl+p", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("ctrl+n", "down"),
			key.WithHelp("Ctrl+n", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select"),
		),
		Delete: key.NewBinding(
			key.WithKeys("delete"),
			key.WithHelp("Delete", "delete"),
		),
	}
}

type ResumeMatchResult struct {
	Item  utils.UnfinishedDownload
	Score float64
}

type ResumeModel struct {
	Visible      bool
	AllItems     []utils.UnfinishedDownload
	Filtered     []ResumeMatchResult
	SelectedIdx  int
	ScrollOffset int
	Query        string
	Keys         ResumeKeyMap
	Width        int
	MaxHeight    int
}

func NewResumeModel() ResumeModel {
	return ResumeModel{
		Visible:      false,
		AllItems:     []utils.UnfinishedDownload{},
		Filtered:     []ResumeMatchResult{},
		SelectedIdx:  0,
		ScrollOffset: 0,
		Query:        "",
		Keys:         DefaultResumeKeyMap(),
		Width:        60,
		MaxHeight:    10,
	}
}

func (m *ResumeModel) Show() {
	m.Visible = true
	m.Query = ""
	m.LoadItems()
}

func (m *ResumeModel) Hide() {
	m.Visible = false
	m.AllItems = []utils.UnfinishedDownload{}
	m.Filtered = []ResumeMatchResult{}
	m.SelectedIdx = 0
	m.ScrollOffset = 0
	m.Query = ""
}

func (m *ResumeModel) LoadItems() {
	items, err := utils.LoadUnfinished()
	if err != nil {
		m.AllItems = []utils.UnfinishedDownload{}
		m.Filtered = []ResumeMatchResult{}
		return
	}

	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if items[i].Timestamp.Before(items[j].Timestamp) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	m.AllItems = items
	m.UpdateFilteredItems()
}

func (m *ResumeModel) UpdateFilteredItems() {
	if m.Query == "" {
		m.Filtered = make([]ResumeMatchResult, len(m.AllItems))
		for i, item := range m.AllItems {
			m.Filtered[i] = ResumeMatchResult{Item: item, Score: 1000}
		}
	} else {
		patterns := make([]string, len(m.AllItems))
		for i, item := range m.AllItems {
			patterns[i] = item.Title + " " + item.URL
		}

		matches := fuzzy.Find(m.Query, patterns)

		m.Filtered = []ResumeMatchResult{}
		addedItems := make(map[int]bool)

		for i, item := range m.AllItems {
			lowerQuery := strings.ToLower(m.Query)
			lowerTitle := strings.ToLower(item.Title)
			lowerURL := strings.ToLower(item.URL)

			if strings.Contains(lowerTitle, lowerQuery) || strings.Contains(lowerURL, lowerQuery) {
				m.Filtered = append(m.Filtered, ResumeMatchResult{Item: item, Score: 1001})
				addedItems[i] = true
			}
		}

		for _, match := range matches {
			if !addedItems[match.Index] {
				m.Filtered = append(m.Filtered, ResumeMatchResult{
					Item:  m.AllItems[match.Index],
					Score: float64(match.Score),
				})
			}
		}
	}

	if m.SelectedIdx >= len(m.Filtered) {
		m.SelectedIdx = 0
	}
}

func (m *ResumeModel) SetQuery(query string) {
	m.Query = query
	m.SelectedIdx = 0
	m.UpdateFilteredItems()
}

func (m *ResumeModel) Next() {
	if len(m.Filtered) == 0 {
		return
	}

	m.SelectedIdx++
	if m.SelectedIdx >= len(m.Filtered) {
		m.SelectedIdx = 0
	}
}

func (m *ResumeModel) Prev() {
	if len(m.Filtered) == 0 {
		return
	}

	m.SelectedIdx--
	if m.SelectedIdx < 0 {
		m.SelectedIdx = len(m.Filtered) - 1
	}
}

func (m *ResumeModel) updateScrollOffset(height int) {
	if len(m.Filtered) == 0 {
		return
	}

	visibleItems := min(height-2, m.MaxHeight)

	if m.ScrollOffset < 0 {
		m.ScrollOffset = 0
	}

	if m.ScrollOffset > len(m.Filtered)-visibleItems {
		m.ScrollOffset = max(0, len(m.Filtered)-visibleItems)
	}

	if m.SelectedIdx < m.ScrollOffset {
		m.ScrollOffset = m.SelectedIdx
	}

	if m.SelectedIdx >= m.ScrollOffset+visibleItems {
		m.ScrollOffset = m.SelectedIdx - visibleItems + 1
	}
}

func (m *ResumeModel) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.Visible {
			return false, nil
		}

		switch {
		case key.Matches(msg, m.Keys.Up):
			m.Prev()
			return true, nil
		case key.Matches(msg, m.Keys.Down):
			m.Next()
			return true, nil
		case key.Matches(msg, m.Keys.Select):
			return true, nil
		case key.Matches(msg, m.Keys.Delete):
			m.DeleteSelected()
			return true, nil
		case msg.Type == tea.KeyRunes:
			m.Query += string(msg.Runes)
			m.SelectedIdx = 0
			m.UpdateFilteredItems()
			return true, nil
		case msg.Type == tea.KeyBackspace:
			if len(m.Query) > 0 {
				m.Query = m.Query[:len(m.Query)-1]
				m.SelectedIdx = 0
				m.UpdateFilteredItems()
			}
			return true, nil
		case msg.Type == tea.KeyEsc:
			m.Hide()
			return true, nil
		}
	}

	return false, nil
}

func (m *ResumeModel) HandleResize(width, height int) {
	m.Width = width - 4
}

func (m *ResumeModel) DeleteSelected() {
	if m.SelectedIdx < 0 || m.SelectedIdx >= len(m.Filtered) {
		return
	}

	url := m.Filtered[m.SelectedIdx].Item.URL
	if err := utils.RemoveUnfinished(url); err != nil {
		return
	}

	m.LoadItems()
}

func (m *ResumeModel) SelectedItem() *utils.UnfinishedDownload {
	if m.SelectedIdx >= 0 && m.SelectedIdx < len(m.Filtered) {
		return &m.Filtered[m.SelectedIdx].Item
	}

	return nil
}

func (m *ResumeModel) View(width, height int) string {
	if !m.Visible || len(m.Filtered) == 0 {
		return ""
	}

	if height < 4 {
		return ""
	}

	m.updateScrollOffset(height)

	var b strings.Builder

	visibleItems := height - 2
	if m.Query != "" {
		visibleItems--
	}
	if visibleItems > m.MaxHeight {
		visibleItems = m.MaxHeight
	}

	if m.ScrollOffset < 0 {
		m.ScrollOffset = 0
	}

	if m.ScrollOffset > len(m.Filtered)-visibleItems {
		m.ScrollOffset = max(0, len(m.Filtered)-visibleItems)
	}

	for i := 0; i < visibleItems; i++ {
		idx := m.ScrollOffset + i
		if idx >= len(m.Filtered) {
			break
		}

		item := m.Filtered[idx].Item
		isSelected := idx == m.SelectedIdx

		title := item.Title
		if len(title) > width-10 {
			title = title[:width-13] + "..."
		}

		url := item.URL
		if len(url) > width-10 {
			url = url[:width-13] + "..."
		}

		var titleStyle string
		if isSelected {
			titleStyle = styles.AutocompleteSelected.Render("> " + title)
		} else {
			titleStyle = styles.AutocompleteItem.Render("  " + title)
		}

		urlStyle := styles.AutocompleteItem.Foreground(styles.MutedColor).Render("  " + url)

		b.WriteRune('\n')
		b.WriteString(titleStyle)
		b.WriteRune('\n')
		b.WriteString(urlStyle)

		if i < visibleItems-1 && idx < len(m.Filtered)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
