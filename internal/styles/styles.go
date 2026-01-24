package styles

import "github.com/charmbracelet/lipgloss"

var (
	PrimaryColor   = lipgloss.Color("#ffffff")
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
				Foreground(PrimaryColor)
	StatusBarStyle = lipgloss.NewStyle().Foreground(MutedColor).MarginTop(1)
	InputStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false).BorderForeground(MutedColor)

	listStyle              = lipgloss.NewStyle().Padding(0, 3)
	ListTitleStyle         = listStyle.Foreground(lipgloss.Color("#bac2de"))
	ListSelectedTitleStyle = listStyle.Foreground(lipgloss.Color("#b4befe")).Bold(true)
	ListDescStyle          = listStyle.Foreground(lipgloss.Color("#6c7086"))
	ListSelectedDescStyle  = listStyle.Foreground(lipgloss.Color("#cdd6f4"))

	SpinnerStyle = lipgloss.NewStyle().Foreground(PinkColor)
)
