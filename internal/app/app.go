package app

import (
	"fmt"
	"strings"
	"xytz/internal/models"
	"xytz/internal/styles"
	"xytz/internal/types"
	"xytz/internal/utils"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Program      *tea.Program
	Search       models.SearchModel
	State        types.State
	Width        int
	Height       int
	Spinner      spinner.Model
	LoadingType  string
	CurrentQuery string
	Videos       []list.Item
	VideoList    models.VideoListModel
	ErrMsg       string
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.Search.Init(), m.Spinner.Tick)
}

func NewModel() *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = s.Style.Foreground(styles.PinkColor)

	return &Model{
		Search:    models.NewSearchModel(),
		State:     types.StateSearchInput,
		Spinner:   s,
		VideoList: models.NewVideoListModel(),
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Search = m.Search.HandleResize(m.Width, m.Height)
		m.VideoList = m.VideoList.HandleResize(m.Width, m.Height)
	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		return m, spinnerCmd
	case types.StartSearchMsg:
		m.State = types.StateLoading
		m.LoadingType = "search"
		m.CurrentQuery = strings.TrimSpace(msg.Query)
		cmd = utils.PerformSearch(msg.Query)
		m.ErrMsg = ""
	case types.SearchResultMsg:
		m.LoadingType = ""
		m.Videos = msg.Videos
		m.VideoList.List.SetItems(msg.Videos)
		m.VideoList.CurrentQuery = m.CurrentQuery
		m.State = types.StateVideoList
		m.ErrMsg = msg.Err
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
		switch m.State {
		case types.StateSearchInput:
			updatedSearch, searchCmd := m.Search.Update(msg)
			m.Search = updatedSearch.(models.SearchModel)
			cmd = searchCmd
		case types.StateVideoList:
			updatedList, listCmd := m.VideoList.Update(msg)
			m.VideoList = updatedList.(models.VideoListModel)
			cmd = listCmd
		}
	}

	return m, cmd
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	var content string
	switch m.State {
	case types.StateSearchInput:
		content = m.Search.View()
	case types.StateLoading:
		content = m.LoadingView()
	case types.StateVideoList:
		content = m.VideoList.View()
	}
	var left string

	switch m.State {

	case types.StateSearchInput:
		left = "Ctrl+C: quit"
	default:
		left = "Ctrl+C: quit • q: quit"
	}

	right := ""
	if m.ErrMsg != "" {
		right = lipgloss.NewStyle().Foreground(styles.ErrorColor).Render(m.ErrMsg)
	}

	var statusBar string
	if right != "" {
		statusBar = styles.StatusBarStyle.Height(1).Width(m.Width).Render(left + lipgloss.PlaceHorizontal(m.Width-lipgloss.Width(left), lipgloss.Right, right))
	} else {
		statusBar = styles.StatusBarStyle.Height(1).Width(m.Width).Render(left)
	}

	return lipgloss.JoinVertical(lipgloss.Top, content, statusBar)
}

func (m *Model) LoadingView() string {
	var s strings.Builder

	loadingText := "Loading..."
	switch m.LoadingType {
	case "search":
		loadingText = fmt.Sprintf("Searching for \"%s\"", m.CurrentQuery)
	}

	fmt.Fprintf(&s, "\n%s %s\n", m.Spinner.View(), loadingText)

	return s.String()
}
