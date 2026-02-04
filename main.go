package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/xdagiz/xytz/internal/app"
	"github.com/xdagiz/xytz/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

func main() {
	zone.NewGlobal()
	defer zone.Close()

	m := app.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.Program = p

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, ".local", "share", "xytz")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Could not create log directory: %v", err)
		logDir = "."
	}

	logPath := filepath.Join(logDir, "debug.log")

	logger, err := tea.LogToFile(logPath, "debug")
	if err != nil {
		log.Printf("Warning: Could not create debug log file: %v", err)
	} else {
		defer logger.Close()
	}

	if _, err := p.Run(); err != nil {
		log.Fatal("unable to run the app")
		os.Exit(1)
	}

	saveConfigOptions(m)
}

func saveConfigOptions(m *app.Model) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config on exit: %v", err)
		return
	}

	for _, opt := range m.Search.DownloadOptions {
		switch opt.ConfigField {
		case "EmbedSubtitles":
			cfg.EmbedSubtitles = opt.Enabled
		case "EmbedMetadata":
			cfg.EmbedMetadata = opt.Enabled
		case "EmbedChapters":
			cfg.EmbedChapters = opt.Enabled
		}
	}

	cfg.SortByDefault = string(m.Search.SortBy)

	if err := cfg.Save(); err != nil {
		log.Printf("Failed to save config on exit: %v", err)
	}
}
