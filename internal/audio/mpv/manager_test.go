package mpv

import (
	"testing"

	"navitone-cli/internal/models"
)

func TestUpdateStreamMetadata(t *testing.T) {
	m := &Manager{
		queue: []models.Track{
			{
				ID:     "rad-1",
				Title:  "Nightride FM",
				Artist: "Radio Stream",
				Album:  "Internet Radio",
				Path:   "https://stream.nightride.fm/nightride.mp3",
			},
		},
		currentIndex: 0,
	}

	// Test updating metadata with "Artist - Title"
	m.updateStreamMetadataLocked("Alpha Chrome Yayo, Danny Madigan - In Time")

	if m.queue[0].Artist != "Alpha Chrome Yayo, Danny Madigan" {
		t.Errorf("Expected artist 'Alpha Chrome Yayo, Danny Madigan', got %q", m.queue[0].Artist)
	}
	if m.queue[0].Title != "In Time" {
		t.Errorf("Expected title 'In Time', got %q", m.queue[0].Title)
	}

	// Test updating metadata with another song
	m.updateStreamMetadataLocked("Windows 96 - Caligula")

	if m.queue[0].Artist != "Windows 96" {
		t.Errorf("Expected artist 'Windows 96', got %q", m.queue[0].Artist)
	}
	if m.queue[0].Title != "Caligula" {
		t.Errorf("Expected title 'Caligula', got %q", m.queue[0].Title)
	}
}
