package utils

import (
	"testing"
)

func TestNormalizeRadioURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nightride FM deprecated m4a",
			input:    "https://stream.nightride.fm/nightride.m4a",
			expected: "https://stream.nightride.fm/nightride.mp3",
		},
		{
			name:     "Nightride FM Datawave m4a",
			input:    "https://stream.nightride.fm/datawave.m4a",
			expected: "https://stream.nightride.fm/datawave.mp3",
		},
		{
			name:     "Nightride FM Datawave no extension",
			input:    "https://stream.nightride.fm/datawave",
			expected: "https://stream.nightride.fm/datawave.mp3",
		},
		{
			name:     "Nightride FM Datawave opus extension",
			input:    "https://stream.nightride.fm/datawave.opus",
			expected: "https://stream.nightride.fm/datawave.mp3",
		},
		{
			name:     "Nightride FM Datawave aac extension",
			input:    "https://stream.nightride.fm/datawave.aac",
			expected: "https://stream.nightride.fm/datawave.mp3",
		},
		{
			name:     "Nightride FM Datawave m4a with query parameters",
			input:    "https://stream.nightride.fm/datawave.m4a?stream=live",
			expected: "https://stream.nightride.fm/datawave.mp3?stream=live",
		},
		{
			name:     "Nightride FM stream subpath m4a",
			input:    "https://nightride.fm/stream/chillsynth.m4a",
			expected: "https://nightride.fm/stream/chillsynth.mp3",
		},
		{
			name:     "Nightride FM stream subpath no extension",
			input:    "https://nightride.fm/stream/datawave",
			expected: "https://nightride.fm/stream/datawave.mp3",
		},
		{
			name:     "Nightride FM already mp3",
			input:    "https://stream.nightride.fm/nightride.mp3",
			expected: "https://stream.nightride.fm/nightride.mp3",
		},
		{
			name:     "Nightride FM already ogg",
			input:    "https://stream.nightride.fm/nightride.ogg",
			expected: "https://stream.nightride.fm/nightride.ogg",
		},
		{
			name:     "SomaFM mp3 stream unchanged",
			input:    "http://ice1.somafm.com/groovesalad-128-mp3",
			expected: "http://ice1.somafm.com/groovesalad-128-mp3",
		},
		{
			name:     "Nightride FM leading tab and space padding",
			input:    "\t https://stream.nightride.fm/datawave.mp3",
			expected: "https://stream.nightride.fm/datawave.mp3",
		},
		{
			name:     "Empty input unchanged",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeRadioURL(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeRadioURL(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}
