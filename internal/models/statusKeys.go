package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/xdagiz/xytz/internal/types"
)

type StatusKeys struct {
	Quit   key.Binding
	Back   key.Binding
	Enter  key.Binding
	Pause  key.Binding
	Cancel key.Binding
	Tab    key.Binding
	Help   key.Binding
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Delete key.Binding
	Next   key.Binding
	Prev   key.Binding
}

func GetStatusKeys(state types.State, helpVisible bool, resumeVisible bool, resumeKeys ResumeKeyMap) StatusKeys {
	keys := StatusKeys{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("Ctrl+c/q", "quit"),
		),
	}

	switch state {
	case types.StateSearchInput:
		if resumeVisible {
			keys.Up = resumeKeys.Up
			keys.Down = resumeKeys.Down
			keys.Select = resumeKeys.Select
			keys.Delete = resumeKeys.Delete
			keys.Cancel = key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("Esc", "cancel"),
			)
		}
	case types.StateVideoList:
		keys.Back = key.NewBinding(
			key.WithKeys("esc", "b"),
			key.WithHelp("Esc/b", "back"),
		)
	case types.StateFormatList:
		keys.Back = key.NewBinding(
			key.WithKeys("esc", "b"),
			key.WithHelp("Esc/b", "back"),
		)
	case types.StateDownload:
		keys.Back = key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "back"),
		)
		keys.Enter = key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "back to search"),
		)
		keys.Pause = key.NewBinding(
			key.WithKeys("p", " "),
			key.WithHelp("p/space", "pause"),
		)
		keys.Cancel = key.NewBinding(
			key.WithKeys("esc", "c"),
			key.WithHelp("Esc/c", "cancel"),
		)
	}

	return keys
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

func FormatKeysForStatusBar(keys StatusKeys) string {
	var parts []string

	addKey := func(binding key.Binding) {
		if text := formatKey(binding, false); text != "" {
			parts = append(parts, text)
		}
	}

	addKey(keys.Quit)
	addKey(keys.Back)
	addKey(keys.Enter)
	addKey(keys.Pause)
	addKey(keys.Cancel)
	addKey(keys.Tab)
	addKey(keys.Help)
	addKey(keys.Up)
	addKey(keys.Down)
	addKey(keys.Select)
	addKey(keys.Delete)
	addKey(keys.Next)
	addKey(keys.Prev)

	return strings.Join(parts, " • ")
}

func FormatKeysForStatusBarItalic(keys StatusKeys, italicKey string) string {
	var parts []string

	addKey := func(binding key.Binding, keyName string) {
		if text := formatKey(binding, keyName == italicKey); text != "" {
			parts = append(parts, text)
		}
	}

	addKey(keys.Quit, "Quit")
	addKey(keys.Back, "Back")
	addKey(keys.Enter, "Enter")
	addKey(keys.Pause, "Pause")
	addKey(keys.Cancel, "Cancel")
	addKey(keys.Tab, "Tab")
	addKey(keys.Help, "Help")
	addKey(keys.Up, "Up")
	addKey(keys.Down, "Down")
	addKey(keys.Select, "Select")
	addKey(keys.Delete, "Delete")

	return strings.Join(parts, " • ")
}

func FormatSingleKey(binding key.Binding) string {
	return formatKey(binding, false)
}
