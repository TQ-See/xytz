package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	t.Run("returns non-empty string", func(t *testing.T) {
		dir := GetConfigDir()
		if dir == "" {
			t.Error("GetConfigDir() returned empty string")
		}
	})

	t.Run("ends with xytz", func(t *testing.T) {
		dir := GetConfigDir()
		expectedSuffix := "xytz"
		if len(dir) < len(expectedSuffix) || dir[len(dir)-len(expectedSuffix):] != expectedSuffix {
			t.Errorf("GetConfigDir() = %q, should end with %q", dir, expectedSuffix)
		}
	})
}

func TestGetDataDir(t *testing.T) {
	t.Run("returns non-empty string", func(t *testing.T) {
		dir := GetDataDir()
		if dir == "" {
			t.Error("GetDataDir() returned empty string")
		}
	})

	t.Run("ends with xytz", func(t *testing.T) {
		dir := GetDataDir()
		expectedSuffix := "xytz"
		if len(dir) < len(expectedSuffix) || dir[len(dir)-len(expectedSuffix):] != expectedSuffix {
			t.Errorf("GetDataDir() = %q, should end with %q", dir, expectedSuffix)
		}
	})
}

func TestEnsureDirExists(t *testing.T) {
	t.Run("creates directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "xytz-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		testDir := tmpDir + "/newdir"
		err = EnsureDirExists(testDir)
		if err != nil {
			t.Errorf("EnsureDirExists() error = %v", err)
		}

		// Verify directory was created
		info, err := os.Stat(testDir)
		if err != nil {
			t.Errorf("Failed to stat directory: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected path to be a directory")
		}
	})

	t.Run("succeeds if directory exists", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "xytz-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		err = EnsureDirExists(tmpDir)
		if err != nil {
			t.Errorf("EnsureDirExists() error = %v", err)
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "xytz-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		nestedDir := tmpDir + "/level1/level2/level3"
		err = EnsureDirExists(nestedDir)
		if err != nil {
			t.Errorf("EnsureDirExists() error = %v", err)
		}

		info, err := os.Stat(nestedDir)
		if err != nil {
			t.Errorf("Failed to stat directory: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected path to be a directory")
		}
	})
}

func TestShortenHome(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	home, err := os.UserHomeDir()
	if err != nil || filepath.Clean(home) != filepath.Clean(tmpHome) {
		t.Skip("cannot override home directory on this platform")
	}
	sep := string(filepath.Separator)

	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"exact home", tmpHome, "~"},
		{"home trailing separator", tmpHome + sep, "~"},
		{"nested file", filepath.Join(tmpHome, "Videos", "a.mp4"), "~" + sep + filepath.Join("Videos", "a.mp4")},
		{"outside home", sep + filepath.Join("tmp", "a.mp4"), sep + filepath.Join("tmp", "a.mp4")},
		{"sibling prefix", tmpHome + "2" + sep + "x.mp4", tmpHome + "2" + sep + "x.mp4"},
		{"already shortened", "~/Videos/a.mp4", "~/Videos/a.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortenHome(tt.path); got != tt.want {
				t.Errorf("ShortenHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestDisplayPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	home, err := os.UserHomeDir()
	if err != nil || filepath.Clean(home) != filepath.Clean(tmpHome) {
		t.Skip("cannot override home directory on this platform")
	}
	sep := string(filepath.Separator)

	if got := DisplayPath(""); got != "" {
		t.Errorf("DisplayPath empty = %q, want empty", got)
	}

	want := "~" + sep + filepath.Join("Music", "a.mp3")
	if got := DisplayPath(filepath.Join(tmpHome, "Music", "a.mp3")); got != want {
		t.Errorf("DisplayPath = %q, want %q", got, want)
	}

	outside := sep + filepath.Join("tmp", "a.mp3")
	if got := DisplayPath(outside); got != outside {
		t.Errorf("DisplayPath outside = %q, want %q", got, outside)
	}
}

func TestTruncatePathTail(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	home, err := os.UserHomeDir()
	if err != nil || filepath.Clean(home) != filepath.Clean(tmpHome) {
		t.Skip("cannot override home directory on this platform")
	}
	sep := string(filepath.Separator)

	short := filepath.Join(tmpHome, "Videos", "short.mp4")
	wantShort := "~" + sep + filepath.Join("Videos", "short.mp4")
	if got := TruncatePathTail(short, 16); got != wantShort {
		t.Errorf("TruncatePathTail short = %q, want %q", got, wantShort)
	}

	longTitle := filepath.Join(tmpHome, "Videos", "averylongvideotitlethatkeepsgoing.mp4")
	got := TruncatePathTail(longTitle, 16)
	if !filepath.IsAbs(got) && got[0] != '~' {
		t.Errorf("TruncatePathTail long = %q, want shortened dir", got)
	}
	if filepath.Ext(got) != ".mp4" {
		t.Errorf("TruncatePathTail long = %q, want .mp4 extension preserved", got)
	}

	if got := TruncatePathTail("", 16); got != "" {
		t.Errorf("TruncatePathTail empty = %q, want empty", got)
	}

	if got := TruncatePathTail(short, 0); got != short {
		t.Errorf("TruncatePathTail zero limit = %q, want unchanged", got)
	}
}

func TestPlatformSpecificPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific tests on Windows")
	}

	if runtime.GOOS == "linux" {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		dir := GetConfigDir()
		if dir != "/custom/config/xytz" {
			t.Errorf("GetConfigDir() = %q, expected XDG_CONFIG_HOME to be respected", dir)
		}

		t.Setenv("XDG_DATA_HOME", "/custom/data")
		dir = GetDataDir()
		if dir != "/custom/data/xytz" {
			t.Errorf("GetDataDir() = %q, expected XDG_DATA_HOME to be respected", dir)
		}
	}

	if runtime.GOOS == "darwin" {
		t.Run("XDG_CONFIG_HOME respected on darwin", func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", "/custom/config")
			dir := GetConfigDir()
			if dir != "/custom/config/xytz" {
				t.Errorf("GetConfigDir() = %q, expected XDG_CONFIG_HOME to be respected on darwin", dir)
			}
		})

		t.Run("XDG_DATA_HOME respected on darwin", func(t *testing.T) {
			t.Setenv("XDG_DATA_HOME", "/custom/data")
			dir := GetDataDir()
			if dir != "/custom/data/xytz" {
				t.Errorf("GetDataDir() = %q, expected XDG_DATA_HOME to be respected on darwin", dir)
			}
		})
	}
}
