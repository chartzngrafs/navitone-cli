package controllers

import (
	"testing"

	"navitone-cli/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchHotkey(t *testing.T) {
	app := &App{
		state: &models.AppState{
			ConfigForm: &models.ConfigFormState{},
		},
	}

	if app.state.ShowSearchModal {
		t.Fatalf("Expected ShowSearchModal to be false initially")
	}

	// Press '/' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	app.handleKeyPress(keyMsg)

	if !app.state.ShowSearchModal {
		t.Errorf("Expected ShowSearchModal to be true after pressing '/', got false")
	}
}

func TestZenModeHotkey(t *testing.T) {
	app := &App{
		state: &models.AppState{
			ConfigForm: &models.ConfigFormState{},
		},
	}

	if app.state.ZenMode {
		t.Fatalf("Expected ZenMode to be false initially")
	}

	// Press 'Shift+Z' key ('Z')
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}}
	app.handleKeyPress(keyMsg)

	if !app.state.ZenMode {
		t.Errorf("Expected ZenMode to be true after pressing 'Z', got false")
	}

	// Press 'Z' key again to disable
	app.handleKeyPress(keyMsg)

	if app.state.ZenMode {
		t.Errorf("Expected ZenMode to be false after toggling again, got true")
	}
}

func TestRadioTabNavigation(t *testing.T) {
	app := &App{
		state: &models.AppState{
			ConfigForm: &models.ConfigFormState{},
			CurrentTab: models.RadioTab,
			RadioStations: []models.RadioStation{
				{ID: "radio-1", Name: "Radio Paradise", StreamURL: "http://stream.radioparadise.com/flac"},
				{ID: "radio-2", Name: "SomaFM Groove Salad", StreamURL: "http://ice1.somafm.com/groovesalad-128-mp3"},
			},
		},
	}

	if app.state.SelectedRadioIndex != 0 {
		t.Fatalf("Expected SelectedRadioIndex to be 0 initially")
	}

	// Press Down arrow key
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	app.handleKeyPress(keyMsg)

	if app.state.SelectedRadioIndex != 1 {
		t.Errorf("Expected SelectedRadioIndex to be 1 after pressing Down arrow, got %d", app.state.SelectedRadioIndex)
	}

	// Press Up arrow key
	keyUpMsg := tea.KeyMsg{Type: tea.KeyUp}
	app.handleKeyPress(keyUpMsg)

	if app.state.SelectedRadioIndex != 0 {
		t.Errorf("Expected SelectedRadioIndex to be 0 after pressing Up arrow, got %d", app.state.SelectedRadioIndex)
	}
}

func TestTabCyclingAll7Tabs(t *testing.T) {
	app := &App{
		state: &models.AppState{
			ConfigForm: &models.ConfigFormState{},
			CurrentTab: models.HomeTab,
		},
	}

	expectedTabs := []models.Tab{
		models.AlbumsTab,
		models.ArtistsTab,
		models.PlaylistsTab,
		models.RadioTab,
		models.QueueTab,
		models.ConfigTab,
		models.HomeTab,
	}

	for _, expected := range expectedTabs {
		app.nextTab()
		if app.state.CurrentTab != expected {
			t.Errorf("Expected next tab %v, got %v", expected, app.state.CurrentTab)
		}
	}
}

func TestNormalizeRadioURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://stream.nightride.fm/nightride.m4a", "https://stream.nightride.fm/nightride.mp3"},
		{"https://nightride.fm/stream/chillsynth.m4a", "https://nightride.fm/stream/chillsynth.mp3"},
		{"https://stream.nightride.fm/datawave.m4a", "https://stream.nightride.fm/datawave.mp3"},
		{"https://stream.nightride.fm/datawave", "https://stream.nightride.fm/datawave.mp3"},
		{"https://stream.nightride.fm/datawave.opus", "https://stream.nightride.fm/datawave.mp3"},
		{"https://stream.nightride.fm/nightride.mp3", "https://stream.nightride.fm/nightride.mp3"},
		{"http://ice1.somafm.com/groovesalad-128-mp3", "http://ice1.somafm.com/groovesalad-128-mp3"},
	}

	for _, tt := range tests {
		got := normalizeRadioURL(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeRadioURL(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
