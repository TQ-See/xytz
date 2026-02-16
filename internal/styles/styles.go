package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	PrimaryColor   = lipgloss.Color("#ffffff")
	BlackColor     = lipgloss.Color("#1e1e2e")
	SecondaryColor = lipgloss.Color("#cdd6f4")
	ErrorColor     = lipgloss.Color("#f38ba8")
	SuccessColor   = lipgloss.Color("#a6e3a1")
	WarningColor   = lipgloss.Color("#f9e2af")
	InfoColor      = lipgloss.Color("#89dceb")
	MutedColor     = lipgloss.Color("#6c7086")
	MaroonColor    = lipgloss.Color("#eba0ac")
	PinkColor      = lipgloss.Color("#f5c2e7")
	MauveColor     = lipgloss.Color("#cba6f7")
)

var (
	ASCIIStyle         = lipgloss.NewStyle().Foreground(MauveColor).PaddingBottom(1)
	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SecondaryColor).
				Padding(1, 0)
	StatusBarStyle = lipgloss.NewStyle().Foreground(MutedColor).Padding(1, 2)
	InputStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderForeground(MutedColor)
	MutedStyle     = lipgloss.NewStyle().Foreground(MutedColor)

	listStyle              = lipgloss.NewStyle().Padding(0, 3)
	ListTitleStyle         = listStyle.Foreground(lipgloss.Color("#bac2de"))
	ListSelectedTitleStyle = listStyle.Foreground(MauveColor).Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(MauveColor).
				Padding(0, 0, 0, 2)

	ListDescStyle         = listStyle.Foreground(MutedColor)
	ListSelectedDescStyle = listStyle.Foreground(SecondaryColor)
	ListDimmedTitle       = listStyle.Foreground(MutedColor).
				Padding(0, 0, 0, 3)
	ListDimmedDesc = listStyle.Foreground(MutedColor)

	ListSelectedQueueStyle = lipgloss.NewStyle().Foreground(PinkColor).Bold(true)
	QueueSelectedItemStyle = lipgloss.NewStyle().Foreground(MauveColor).Bold(true)

	ListContainer = lipgloss.NewStyle().PaddingBottom(1)

	SpinnerStyle = lipgloss.NewStyle().Foreground(PinkColor)

	ProgressContainer = lipgloss.NewStyle().PaddingBottom(1)

	SpeedStyle             = lipgloss.NewStyle().Foreground(SuccessColor).Italic(true)
	TimeRemainingStyle     = lipgloss.NewStyle().Foreground(SuccessColor).Italic(true)
	ProgressStyle          = lipgloss.NewStyle().Foreground(SecondaryColor)
	DestinationStyle       = lipgloss.NewStyle().Foreground(MutedColor)
	CompletionMessageStyle = lipgloss.NewStyle().Foreground(SuccessColor)
	HelpStyle              = lipgloss.NewStyle().Foreground(MutedColor).Faint(true)
	ErrorMessageStyle      = lipgloss.NewStyle().Foreground(ErrorColor)
	WarningMessageStyle    = lipgloss.NewStyle().Foreground(WarningColor)

	autocompleteStyle = lipgloss.NewStyle().PaddingLeft(1)
	AutocompleteItem  = autocompleteStyle.
				Foreground(SecondaryColor)
	AutocompleteSelected = autocompleteStyle.
				Foreground(MauveColor)

	sortStyle = lipgloss.NewStyle().PaddingLeft(1)
	SortTitle = sortStyle.Foreground(SecondaryColor).PaddingTop(1).Bold(true)
	SortHelp  = sortStyle.Foreground(MutedColor).Italic(true)
	SortItem  = sortStyle.Foreground(MauveColor).PaddingLeft(1).Italic(true)

	TabActiveStyle   = lipgloss.NewStyle().Foreground(BlackColor).Background(MauveColor)
	TabInactiveStyle = lipgloss.NewStyle().Foreground(SecondaryColor)

	FormatContainerStyle       = lipgloss.NewStyle().PaddingLeft(1)
	CustomFormatContainerStyle = FormatContainerStyle.PaddingLeft(3)
	FormatTabHelpStyle         = lipgloss.NewStyle().Foreground(MutedColor)
	FormatCustomInputStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderForeground(MutedColor).MarginTop(1)
	FormatCustomInputPrompt    = lipgloss.NewStyle().Foreground(PinkColor)
	FormatCustomHelpStyle      = lipgloss.NewStyle().Foreground(MutedColor).PaddingTop(1)
)

func NewListDelegate() list.DefaultDelegate {
	dl := list.NewDefaultDelegate()
	dl.Styles.NormalTitle = ListTitleStyle
	dl.Styles.SelectedTitle = ListSelectedTitleStyle
	dl.Styles.NormalDesc = ListDescStyle
	dl.Styles.SelectedDesc = ListSelectedDescStyle
	dl.Styles.DimmedTitle = ListDimmedTitle
	dl.Styles.DimmedDesc = ListDimmedDesc

	return dl
}
