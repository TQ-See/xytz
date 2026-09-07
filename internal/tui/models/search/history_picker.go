package search

import (
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/xdagiz/xytz/internal/styles"
)

const historyPickerMaxRows = 8

type HistoryMatch struct {
	Text  string
	Score float64
}

type HistoryPicker struct {
	Visible      bool
	Filtered     []HistoryMatch
	SelectedIdx  int
	ScrollOffset int
	MaxRows      int
	Width        int
	query        string
	snapshot     []string
	restoreValue string
	styles       styles.Styles
}

func NewHistoryPicker(st styles.Styles) HistoryPicker {
	return HistoryPicker{
		MaxRows: historyPickerMaxRows,
		Width:   60,
		styles:  st,
	}
}

func (m *HistoryPicker) SetStyles(st styles.Styles) {
	m.styles = st
}

func (m *HistoryPicker) Show(items []string, currentValue string) {
	m.snapshot = append([]string{}, items...)
	m.restoreValue = currentValue
	m.Visible = true
	m.UpdateFilter(currentValue)
}

func (m *HistoryPicker) Hide() {
	m.Visible = false
	m.Filtered = nil
	m.SelectedIdx = 0
	m.ScrollOffset = 0
	m.query = ""
	m.snapshot = nil
	m.restoreValue = ""
}

func (m *HistoryPicker) RestoreValue() string {
	return m.restoreValue
}

func (m *HistoryPicker) Total() int {
	return len(m.snapshot)
}

func (m *HistoryPicker) UpdateFilter(query string) {
	m.query = query
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		m.Filtered = make([]HistoryMatch, len(m.snapshot))
		for i, text := range m.snapshot {
			m.Filtered[i] = HistoryMatch{Text: text}
		}
	} else {
		patterns := make([]string, len(m.snapshot))
		copy(patterns, m.snapshot)
		m.Filtered = nil
		for _, match := range fuzzy.Find(trimmed, patterns) {
			m.Filtered = append(m.Filtered, HistoryMatch{
				Text:  m.snapshot[match.Index],
				Score: float64(match.Score),
			})
		}
	}
	m.SelectedIdx = 0
	m.ScrollOffset = 0
	m.syncScroll()
}

func (m *HistoryPicker) Next() {
	if len(m.Filtered) == 0 {
		return
	}
	m.SelectedIdx++
	if m.SelectedIdx >= len(m.Filtered) {
		m.SelectedIdx = 0
	}
	m.syncScroll()
}

func (m *HistoryPicker) Prev() {
	if len(m.Filtered) == 0 {
		return
	}
	m.SelectedIdx--
	if m.SelectedIdx < 0 {
		m.SelectedIdx = len(m.Filtered) - 1
	}
	m.syncScroll()
}

func (m *HistoryPicker) Selected() string {
	if m.SelectedIdx < 0 || m.SelectedIdx >= len(m.Filtered) {
		return ""
	}
	return m.Filtered[m.SelectedIdx].Text
}

func (m *HistoryPicker) visibleRows() int {
	rows := m.MaxRows
	if rows <= 0 {
		rows = historyPickerMaxRows
	}
	if len(m.Filtered) < rows {
		return len(m.Filtered)
	}
	return rows
}

func (m *HistoryPicker) maxScrollOffset() int {
	maxOffset := len(m.Filtered) - m.visibleRows()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *HistoryPicker) syncScroll() {
	if len(m.Filtered) == 0 {
		m.SelectedIdx = 0
		m.ScrollOffset = 0
		return
	}
	if m.SelectedIdx >= len(m.Filtered) {
		m.SelectedIdx = len(m.Filtered) - 1
	}
	if m.SelectedIdx < 0 {
		m.SelectedIdx = 0
	}
	if m.ScrollOffset > m.maxScrollOffset() {
		m.ScrollOffset = m.maxScrollOffset()
	}
	if m.ScrollOffset < 0 {
		m.ScrollOffset = 0
	}
	visible := m.visibleRows()
	if m.SelectedIdx < m.ScrollOffset {
		m.ScrollOffset = m.SelectedIdx
	}
	if m.SelectedIdx >= m.ScrollOffset+visible {
		m.ScrollOffset = m.SelectedIdx - visible + 1
	}
}

func (m *HistoryPicker) HandleResize(width, height int) {
	m.Width = max(0, width-4)
}

func truncateHistoryText(text string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	if maxChars == 1 {
		return "…"
	}
	return string(runes[:maxChars-1]) + "…"
}

func (m *HistoryPicker) View() string {
	if !m.Visible {
		return ""
	}

	if len(m.snapshot) == 0 {
		return m.styles.MutedStyle.Render("  No history yet. Searches you run will show up here.")
	}

	if len(m.Filtered) == 0 {
		return m.styles.MutedStyle.Render("  No history matches '" + m.query + "'.")
	}

	visible := m.visibleRows()
	end := min(m.ScrollOffset+visible, len(m.Filtered))

	lines := make([]string, 0, visible+1)
	for idx := m.ScrollOffset; idx < end; idx++ {
		text := truncateHistoryText(m.Filtered[idx].Text, m.Width-4)
		if idx == m.SelectedIdx {
			lines = append(lines, m.styles.AutocompleteSelected.Render("› "+text))
		} else {
			lines = append(lines, m.styles.AutocompleteItem.Render("  "+text))
		}
	}
	if end < len(m.Filtered) {
		lines = append(lines, m.styles.MutedStyle.Render("  ↓ more"))
	}

	return strings.Join(lines, "\n")
}
