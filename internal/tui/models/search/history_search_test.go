package search

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func seedHistory(m Model, queries ...string) Model {
	for i := len(queries) - 1; i >= 0; i-- {
		m.History.AddLocal(queries[i])
	}
	return m
}

func TestHistoryPickerCtrlROpensPicker(t *testing.T) {
	setupModelTestEnv(t)

	m := NewModel(testAppCtx(t))
	m = seedHistory(m, "lofi mix", "cooking show", "gaming highlights")

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = updated

	if !m.HistoryPicker.Visible {
		t.Fatalf("expected history picker to open on ctrl+r")
	}
	if got := m.HistoryPicker.Total(); got != 3 {
		t.Fatalf("picker total = %d, want 3", got)
	}
}

func TestHistoryPickerEnterFillsWithoutSubmitting(t *testing.T) {
	setupModelTestEnv(t)

	m := NewModel(testAppCtx(t))
	m = seedHistory(m, "newest", "oldest")

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = updated
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated

	if m.HistoryPicker.Visible {
		t.Fatalf("expected picker to close on enter")
	}
	if got := m.Input.Value(); got != "oldest" {
		t.Fatalf("Input = %q, want %q", got, "oldest")
	}
	if cmd != nil {
		t.Fatalf("expected nil command (fill, not submit)")
	}
}

func TestHistoryPickerEscRestoresInput(t *testing.T) {
	setupModelTestEnv(t)

	m := NewModel(testAppCtx(t))
	m = seedHistory(m, "old query", "older query")
	m.Input.SetValue("half typed")

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = updated
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated

	if m.HistoryPicker.Visible {
		t.Fatalf("expected picker to close on esc")
	}
	if got := m.Input.Value(); got != "half typed" {
		t.Fatalf("Input = %q, want restored %q", got, "half typed")
	}
}

func TestHistoryPickerTypingFilters(t *testing.T) {
	setupModelTestEnv(t)

	m := NewModel(testAppCtx(t))
	m = seedHistory(m, "lofi hip hop", "logic pro tutorial", "cooking pasta")

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	m = updated
	for _, r := range "lofi" {
		updated, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated
	}

	if len(m.HistoryPicker.Filtered) != 1 {
		t.Fatalf("Filtered len = %d, want 1", len(m.HistoryPicker.Filtered))
	}
	if got := m.HistoryPicker.Selected(); got != "lofi hip hop" {
		t.Fatalf("Selected = %q, want %q", got, "lofi hip hop")
	}
}
