package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatalf("DefaultConfig returned nil")
	}

	if cfg.UI.Theme != "dark" {
		t.Errorf("Expected default UI theme 'dark', got '%s'", cfg.UI.Theme)
	}

	if cfg.UI.VimMode != false {
		t.Errorf("Expected default UI VimMode false, got %v", cfg.UI.VimMode)
	}

	if cfg.UI.HomeAlbumCount != 8 {
		t.Errorf("Expected default HomeAlbumCount 8, got %d", cfg.UI.HomeAlbumCount)
	}

	if cfg.Audio.Volume != 100 {
		t.Errorf("Expected default volume 100, got %d", cfg.Audio.Volume)
	}

	if cfg.Navidrome.Timeout != 30 {
		t.Errorf("Expected default Navidrome timeout 30, got %d", cfg.Navidrome.Timeout)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath returned error: %v", err)
	}
	if path == "" {
		t.Errorf("GetConfigPath returned empty string")
	}
}
