package models

import (
	"fmt"
	"strings"

	"github.com/xdagiz/xytz/internal/styles"
	"github.com/xdagiz/xytz/internal/types"
	"github.com/xdagiz/xytz/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
)

type PlayerModel struct {
	URL   string
	Video types.VideoItem
}

func NewPlayer() PlayerModel {
	return PlayerModel{}
}

func (m PlayerModel) Init() tea.Cmd {
	return nil
}

func (m PlayerModel) Update(msg tea.Msg) (PlayerModel, tea.Cmd) {
	return m, nil
}

func (m PlayerModel) View() string {
	var s strings.Builder

	s.WriteString(styles.SectionHeaderStyle.Render("Now Playing"))

	if m.Video.ID != "" {
		s.WriteString(styles.SectionHeaderStyle.Render(m.Video.Title()))
		s.WriteRune('\n')
		s.WriteString(styles.MutedStyle.Render(fmt.Sprintf("⏱  %s", utils.FormatDuration(m.Video.Duration))))
		s.WriteRune('\n')
		s.WriteString(styles.MutedStyle.Render(fmt.Sprintf("👁  %s views", utils.FormatNumber(m.Video.Views))))
		s.WriteRune('\n')
		s.WriteString(styles.MutedStyle.Render(fmt.Sprintf("📺 %s", m.Video.Channel)))
	} else {
		s.WriteString(styles.MutedStyle.Render("No video selected"))
	}

	return s.String()
}
