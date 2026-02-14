package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xdagiz/xytz/internal/paths"
	"gopkg.in/yaml.v3"
)

const ConfigFileName = "config.yaml"

type Config struct {
	SearchLimit         int    `yaml:"search_limit"`
	DefaultDownloadPath string `yaml:"default_download_path"`
	DefaultQuality      string `yaml:"default_quality"`
	SortByDefault       string `yaml:"sort_by_default"`
	EmbedSubtitles      bool   `yaml:"embed_subtitles"`
	EmbedMetadata       bool   `yaml:"embed_metadata"`
	EmbedChapters       bool   `yaml:"embed_chapters"`
	FFmpegPath          string `yaml:"ffmpeg_path"`
	YTDLPPath           string `yaml:"yt_dlp_path"`
	VideoFormat         string `yaml:"video_format"`
	AudioFormat         string `yaml:"audio_format"`
	CookiesBrowser      string `yaml:"cookies_browser"`
	CookiesFile         string `yaml:"cookies_file"`
}

var GetConfigDir = func() string {
	return paths.GetConfigDir()
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), ConfigFileName)
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultCfg := GetDefault()
		if err := defaultCfg.Save(); err != nil {
			return defaultCfg, err
		}

		return defaultCfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: Could not read config file %s: %v, using defaults", configPath, err)
		return GetDefault(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("Warning: Could not parse config file %s: %v, using defaults", configPath, err)
		return GetDefault(), nil
	}

	cfg.applyDefaults()

	return &cfg, nil
}

func (c *Config) Save() error {
	configPath := GetConfigPath()

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0o644)
}

func (c *Config) applyDefaults() {
	defaults := GetDefault()
	if c.SearchLimit == 0 {
		c.SearchLimit = defaults.SearchLimit
	}

	if c.DefaultDownloadPath == "" {
		c.DefaultDownloadPath = defaults.DefaultDownloadPath
	}

	if c.DefaultQuality == "" {
		c.DefaultQuality = defaults.DefaultQuality
	}

	if c.SortByDefault == "" {
		c.SortByDefault = defaults.SortByDefault
	}

	if c.VideoFormat == "" {
		c.VideoFormat = defaults.VideoFormat
	}

	if c.AudioFormat == "" {
		c.AudioFormat = defaults.AudioFormat
	}
}

func (c *Config) GetDefaultFormat() string {
	return ResolveQuality(c.DefaultQuality)
}

func (c *Config) ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}

	return path
}

func (c *Config) GetDownloadPath() string {
	return c.ExpandPath(c.DefaultDownloadPath)
}
