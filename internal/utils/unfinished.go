package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xdagiz/xytz/internal/paths"
	"github.com/xdagiz/xytz/internal/types"
)

var ErrInvalidUnfinishedDownload = errors.New("unfinished download must have valid URL and title")

const UnfinishedFileName = ".xytz_unfinished.json"

type UnfinishedDownload struct {
	URL       string            `json:"url"`
	FormatID  string            `json:"format_id"`
	Title     string            `json:"title"`
	Desc      string            `json:"desc,omitempty"`
	URLs      []string          `json:"urls,omitempty"`
	Videos    []types.VideoItem `json:"videos,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

var GetUnfinishedFilePath = func() string {
	dataDir := paths.GetDataDir()
	if err := paths.EnsureDirExists(dataDir); err != nil {
		log.Printf("Warning: Could not create data directory: %v", err)
		return UnfinishedFileName
	}

	return filepath.Join(dataDir, UnfinishedFileName)
}

func LoadUnfinished() ([]UnfinishedDownload, error) {
	path := GetUnfinishedFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []UnfinishedDownload{}, nil
		}

		return nil, err
	}

	if len(data) == 0 {
		return []UnfinishedDownload{}, nil
	}

	var downloads []UnfinishedDownload
	if err := json.Unmarshal(data, &downloads); err != nil {
		return nil, err
	}

	return downloads, nil
}

func SaveUnfinished(downloads []UnfinishedDownload) error {
	if downloads == nil {
		downloads = []UnfinishedDownload{}
	}

	path := GetUnfinishedFilePath()
	data, err := json.MarshalIndent(downloads, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func AddUnfinished(download UnfinishedDownload) error {
	if download.URL == "" || download.Title == "" {
		return ErrInvalidUnfinishedDownload
	}

	downloads, err := LoadUnfinished()
	if err != nil {
		return err
	}

	for i, d := range downloads {
		if d.URL == download.URL {
			downloads[i] = download
			return SaveUnfinished(downloads)
		}
	}

	downloads = append(downloads, download)
	return SaveUnfinished(downloads)
}

func RemoveUnfinished(url string) error {
	downloads, err := LoadUnfinished()
	if err != nil {
		return err
	}

	var newDownloads []UnfinishedDownload
	for _, d := range downloads {
		if d.URL != url {
			newDownloads = append(newDownloads, d)
		}
	}

	return SaveUnfinished(newDownloads)
}

func GetUnfinishedByURL(url string) *UnfinishedDownload {
	downloads, err := LoadUnfinished()
	if err != nil {
		return nil
	}

	for _, d := range downloads {
		if d.URL == url {
			return &d
		}
	}

	return nil
}

func AddUnfinishedBatch(downloads []UnfinishedDownload) error {
	if len(downloads) == 0 {
		return nil
	}

	existing, err := LoadUnfinished()
	if err != nil {
		return err
	}

	existingMap := make(map[string]int)
	for i, d := range existing {
		existingMap[d.URL] = i
	}

	for _, d := range downloads {
		if d.URL == "" || d.Title == "" {
			continue
		}

		if idx, exists := existingMap[d.URL]; exists {
			existing[idx] = d
		} else {
			existing = append(existing, d)
		}
	}

	return SaveUnfinished(existing)
}

func RemoveUnfinishedBatch(urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	downloads, err := LoadUnfinished()
	if err != nil {
		return err
	}

	urlSet := make(map[string]bool)
	for _, url := range urls {
		urlSet[url] = true
	}

	var newDownloads []UnfinishedDownload
	for _, d := range downloads {
		if !urlSet[d.URL] {
			newDownloads = append(newDownloads, d)
		}
	}

	return SaveUnfinished(newDownloads)
}

func QueueUnfinishedKey(query string) string {
	return "queue:" + strings.TrimSpace(query)
}
