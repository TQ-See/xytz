package search

import (
	"fmt"
	"strings"
	"testing"
)

func numberedHistoryItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("video %02d lofi mix", i)
	}
	return items
}

func TestHistoryPickerFiltering(t *testing.T) {
	setupModelTestEnv(t)

	p := NewHistoryPicker(testAppCtx(t).Styles)
	p.Show([]string{
		"lofi hip hop beats to relax to",
		"logic pro mixing tutorial",
		"cooking pasta at home",
		"lofi girl livestream",
	}, "")

	for i, want := range []string{
		"lofi hip hop beats to relax to",
		"logic pro mixing tutorial",
		"cooking pasta at home",
		"lofi girl livestream",
	} {
		if p.Filtered[i].Text != want {
			t.Fatalf("Filtered[%d] = %q, want %q", i, p.Filtered[i].Text, want)
		}
	}

	p.UpdateFilter("lofi")
	if len(p.Filtered) != 2 {
		t.Fatalf("Filtered len = %d, want 2", len(p.Filtered))
	}

	p.UpdateFilter("l")
	if len(p.Filtered) != 3 {
		t.Fatalf("single letter Filtered len = %d, want 3", len(p.Filtered))
	}
	for _, match := range p.Filtered {
		if !strings.Contains(strings.ToLower(match.Text), "l") {
			t.Fatalf("unexpected match %q", match.Text)
		}
	}

	p.UpdateFilter("L")
	if len(p.Filtered) != 3 {
		t.Fatalf("uppercase Filtered len = %d, want 3", len(p.Filtered))
	}

	p.UpdateFilter("zzz no such query")
	if len(p.Filtered) != 0 {
		t.Fatalf("Filtered len = %d, want 0", len(p.Filtered))
	}
	if got := p.Selected(); got != "" {
		t.Fatalf("Selected = %q, want empty with no matches", got)
	}
}

func TestHistoryPickerScrollBoundaries(t *testing.T) {
	setupModelTestEnv(t)

	p := NewHistoryPicker(testAppCtx(t).Styles)
	items := numberedHistoryItems(20)
	items[18] = "unique zebra documentary"
	items[19] = "another zebra fact"
	p.Show(items, "")

	for range 19 {
		p.Next()
	}
	if p.SelectedIdx != 19 || p.ScrollOffset != 12 {
		t.Fatalf("bottom: idx=%d offset=%d, want 19 and 12", p.SelectedIdx, p.ScrollOffset)
	}
	if got := p.Selected(); got != "another zebra fact" {
		t.Fatalf("bottom selected = %q", got)
	}
	if lines := strings.Count(p.View(), "\n") + 1; lines != 8 {
		t.Fatalf("bottom lines = %d, want 8 with no markers", lines)
	}

	p.Next()
	if p.SelectedIdx != 0 || p.ScrollOffset != 0 {
		t.Fatalf("wrap forward: idx=%d offset=%d, want 0 and 0", p.SelectedIdx, p.ScrollOffset)
	}

	p.Prev()
	if p.SelectedIdx != 19 || p.ScrollOffset != 12 {
		t.Fatalf("wrap backward: idx=%d offset=%d, want 19 and 12", p.SelectedIdx, p.ScrollOffset)
	}

	for range 19 {
		p.Prev()
	}
	if p.SelectedIdx != 0 || p.ScrollOffset != 0 {
		t.Fatalf("top: idx=%d offset=%d, want 0 and 0", p.SelectedIdx, p.ScrollOffset)
	}
	if lines := strings.Count(p.View(), "\n") + 1; lines != 9 {
		t.Fatalf("top lines = %d, want 9 (8 rows + down marker)", lines)
	}

	for range 19 {
		p.Next()
	}
	p.UpdateFilter("zebra")
	if p.SelectedIdx != 0 || p.ScrollOffset != 0 || len(p.Filtered) != 2 {
		t.Fatalf("after shrink: idx=%d offset=%d matches=%d, want 0, 0 and 2",
			p.SelectedIdx, p.ScrollOffset, len(p.Filtered))
	}
}

func TestHistoryPickerTruncatesNarrowWidth(t *testing.T) {
	setupModelTestEnv(t)

	p := NewHistoryPicker(testAppCtx(t).Styles)
	p.Show([]string{"a very long history entry that exceeds the width"}, "")
	p.HandleResize(14, 40)

	if view := p.View(); !strings.Contains(view, "…") {
		t.Fatalf("narrow view should truncate with ellipsis:\n%s", view)
	}

	p.HandleResize(0, 0)
	if got := p.View(); got == "" {
		t.Fatalf("zero width should still render the prompt")
	}
}
