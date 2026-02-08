package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/xdagiz/xytz/internal/app"
	"github.com/xdagiz/xytz/internal/config"
	"github.com/xdagiz/xytz/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"
)

var (
	searchLimit int
	sortBy      string
	query       string
	channel     string
	playlist    string

	rootCmd = &cobra.Command{
		Use:   "xytz",
		Short: "xytz - YouTube from your terminal",
		Long: `xytz is a TUI YouTube app that allows you to search,
browse, and download videos directly from your terminal.`,
		Run: func(cmd *cobra.Command, args []string) {
			helpFlag, err := cmd.Flags().GetBool("help")
			if err != nil {
				log.Printf("Error getting help flag: %v", err)
				os.Exit(1)
			}

			if helpFlag {
				cmd.Help()
				return
			}

			startApp()
		},
	}
)

func startApp() {
	opts := &models.CLIOptions{
		SearchLimit: searchLimit,
		SortBy:      sortBy,
		Query:       query,
		Channel:     channel,
		Playlist:    playlist,
	}

	zone.NewGlobal()
	defer zone.Close()

	m := app.NewModelWithOptions(opts)
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.GetDefault()
	}

	rootCmd.Flags().IntVarP(&searchLimit, "number", "n", cfg.SearchLimit, "Number of search results")

	rootCmd.Flags().StringVarP(&sortBy, "sort-by", "s", cfg.SortByDefault, "Default sort option (relevance, date, views, rating)")

	rootCmd.Flags().BoolP("help", "h", false, "help for xytz")

	rootCmd.Flags().StringVarP(&query, "query", "q", "", "Direct search with a query")
	rootCmd.Flags().StringVarP(&channel, "channel", "c", "", "Direct channel search")
	rootCmd.Flags().StringVarP(&playlist, "playlist", "p", "", "Direct playlist search")
}
