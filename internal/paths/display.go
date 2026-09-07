package paths

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xdagiz/xytz/internal/utils"
)

func ShortenHome(path string) string {
	if path == "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	absPath := path
	if !filepath.IsAbs(absPath) {
		resolved, err := filepath.Abs(absPath)
		if err != nil {
			return path
		}
		absPath = resolved
	} else {
		absPath = filepath.Clean(absPath)
	}
	rel, err := filepath.Rel(filepath.Clean(home), absPath)
	if err != nil {
		return path
	}
	if rel == "." {
		return "~"
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return "~" + string(filepath.Separator) + rel
}

func DisplayPath(path string) string {
	if path == "" {
		return path
	}
	return ShortenHome(filepath.Clean(path))
}

func TruncatePathTail(path string, maxTitleRunes int) string {
	if path == "" || maxTitleRunes <= 0 {
		return path
	}
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	title := strings.TrimSuffix(base, ext)
	if utf8.RuneCountInString(title) <= maxTitleRunes {
		return ShortenHome(path)
	}
	truncated := utils.Truncate(title, maxTitleRunes+3) + ext
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return truncated
	}
	return filepath.Join(ShortenHome(dir), truncated)
}
