package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/xdagiz/xytz/internal/types"
)

type StatusKeys struct {
	Quit            key.Binding
	Back            key.Binding
	Enter           key.Binding
	Pause           key.Binding
	Cancel          key.Binding
	Tab             key.Binding
	Help            key.Binding
	Up              key.Binding
	Down            key.Binding
	Select          key.Binding
	Delete          key.Binding
	Next            key.Binding
	Prev            key.Binding
	DownloadDefault key.Binding
	SelectVideos    key.Binding
	SelectAll       key.Binding
}

func newQuitKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("Ctrl+c/q", "quit"),
	)
}

func newQuitCtrlCKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("Ctrl+c", "quit"),
	)
}

func newBackEscBKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("esc", "b"),
		key.WithHelp("Esc/b", "back"),
	)
}

func newBackBKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "back"),
	)
}

func newEnterBackToSearchKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "back to search"),
	)
}

func newPauseKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("p", " "),
		key.WithHelp("p/space", "pause"),
	)
}

func newCancelEscKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	)
}

func newCancelEscCKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("esc", "c"),
		key.WithHelp("Esc/c", "cancel"),
	)
}

func newDeleteKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("delete", "ctrl+d"),
		key.WithHelp("Del/Ctrl+d", "delete"),
	)
}

func newDownloadDefaultKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "Download"),
	)
}

func newSelectVideosKey() key.Binding {
	return key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("Space", "select"),
	)
}

func newSelectAllKey() key.Binding {
	return key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	)
}

func GetStatusKeys(state types.State, resumeVisible bool) StatusKeys {
	keys := StatusKeys{
		Quit: newQuitKey(),
	}

	switch state {
	case types.StateSearchInput:
		keys.Quit = newQuitCtrlCKey()
		if resumeVisible {
			keys.Cancel = newCancelEscKey()
			keys.Delete = newDeleteKey()
		}

	case types.StateVideoList:
		keys.Back = newBackEscBKey()
		keys.DownloadDefault = newDownloadDefaultKey()
		keys.SelectVideos = newSelectVideosKey()
		keys.SelectAll = newSelectAllKey()

	case types.StateFormatList:
		keys.Back = newBackEscBKey()

	case types.StateDownload:
		keys.Back = newBackBKey()
		keys.Enter = newEnterBackToSearchKey()
		keys.Pause = newPauseKey()
		keys.Cancel = newCancelEscCKey()
	}

	return keys
}

func LoadingStatusKeys(base StatusKeys) StatusKeys {
	return StatusKeys{
		Quit:   base.Quit,
		Cancel: newCancelEscCKey(),
	}
}

func SearchHelpStatusKeys(helpKeys HelpKeys) StatusKeys {
	return StatusKeys{
		Cancel: newCancelEscKey(),
		Next:   helpKeys.Next,
		Prev:   helpKeys.Prev,
	}
}

func formatKey(binding key.Binding, italic bool) string {
	help := binding.Help()
	if help.Desc == "" && help.Key == "" {
		return ""
	}

	text := help.Key
	if help.Key != "" && help.Desc != "" {
		text = help.Key + ": " + help.Desc
	} else if help.Desc != "" {
		text = help.Desc
	}

	if italic {
		text = lipgloss.NewStyle().Italic(true).Render(help.Key)
		if help.Desc != "" {
			text += ": " + help.Desc
		}
	}

	return text
}

type statusKeyField struct {
	name    string
	binding key.Binding
}

func orderedStatusFields(keys StatusKeys) []statusKeyField {
	return []statusKeyField{
		{name: "Quit", binding: keys.Quit},
		{name: "Back", binding: keys.Back},
		{name: "Enter", binding: keys.Enter},
		{name: "Pause", binding: keys.Pause},
		{name: "Cancel", binding: keys.Cancel},
		{name: "Tab", binding: keys.Tab},
		{name: "Help", binding: keys.Help},
		{name: "Up", binding: keys.Up},
		{name: "Down", binding: keys.Down},
		{name: "Select", binding: keys.Select},
		{name: "Delete", binding: keys.Delete},
		{name: "Next", binding: keys.Next},
		{name: "Prev", binding: keys.Prev},
		{name: "DownloadDefault", binding: keys.DownloadDefault},
		{name: "SelectVideos", binding: keys.SelectVideos},
		{name: "SelectAll", binding: keys.SelectAll},
	}
}

func formatStatusBarKeys(keys StatusKeys, italicKey string) string {
	var parts []string

	for _, field := range orderedStatusFields(keys) {
		if text := formatKey(field.binding, field.name == italicKey); text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " • ")
}

func FormatKeysForStatusBar(keys StatusKeys) string {
	return formatStatusBarKeys(keys, "")
}

func FormatKeysForStatusBarItalic(keys StatusKeys, italicKey string) string {
	return formatStatusBarKeys(keys, italicKey)
}

func FormatSingleKey(binding key.Binding) string {
	return formatKey(binding, false)
}
