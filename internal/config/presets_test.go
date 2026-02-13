package config

import (
	"slices"
	"testing"
)

func TestGetPresetByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *QualityPreset
	}{
		{
			name:  "best preset",
			input: "best",
			expected: &QualityPreset{
				Name:   "best",
				Format: "bv*+ba/b",
			},
		},
		{
			name:  "4k preset",
			input: "4k",
			expected: &QualityPreset{
				Name:   "4k",
				Format: "bv[height<=2160]+ba/b[height<=2160]",
			},
		},
		{
			name:  "1080p preset",
			input: "1080p",
			expected: &QualityPreset{
				Name:   "1080p",
				Format: "bv[height<=1080]+ba/b[height<=1080]",
			},
		},
		{
			name:  "720p preset",
			input: "720p",
			expected: &QualityPreset{
				Name:   "720p",
				Format: "bv[height<=720]+ba/b[height<=720]",
			},
		},
		{
			name:     "invalid preset",
			input:    "invalid",
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPresetByName(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetPresetByName(%q) = %v, want nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("GetPresetByName(%q) = nil, want %v", tt.input, *tt.expected)
				} else if result.Name != tt.expected.Name || result.Format != tt.expected.Format {
					t.Errorf("GetPresetByName(%q) = %+v, want %+v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestIsValidPreset(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "best is valid",
			input:    "best",
			expected: true,
		},
		{
			name:     "4k is valid",
			input:    "4k",
			expected: true,
		},
		{
			name:     "1080p is valid",
			input:    "1080p",
			expected: true,
		},
		{
			name:     "invalid is not valid",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "empty is not valid",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPreset(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidPreset(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveQuality(t *testing.T) {
	tests := []struct {
		name     string
		quality  string
		expected string
	}{
		{
			name:     "empty returns best format",
			quality:  "",
			expected: "bv*+ba/b",
		},
		{
			name:     "preset returns format",
			quality:  "1080p",
			expected: "bv[height<=1080]+ba/b[height<=1080]",
		},
		{
			name:     "unknown returns as-is",
			quality:  "custom-format",
			expected: "custom-format",
		},
		{
			name:     "best returns best format",
			quality:  "best",
			expected: "bv*+ba/b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveQuality(tt.quality)
			if result != tt.expected {
				t.Errorf("ResolveQuality(%q) = %q, want %q", tt.quality, result, tt.expected)
			}
		})
	}
}

func TestPresetNames(t *testing.T) {
	names := PresetNames()

	if len(names) == 0 {
		t.Error("PresetNames() returned empty slice")
	}

	expectedPresets := []string{"best", "4k", "2k", "1080p", "720p", "480p", "360p"}
	if len(names) != len(expectedPresets) {
		t.Errorf("PresetNames() returned %d presets, want %d", len(names), len(expectedPresets))
	}

	for _, expected := range expectedPresets {
		found := slices.Contains(names, expected)

		if !found {
			t.Errorf("Expected preset %q not found in PresetNames()", expected)
		}
	}
}
