package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Navidrome  NavidromeConfig  `toml:"navidrome"`
	Audio      AudioConfig      `toml:"audio"`
	UI         UIConfig         `toml:"ui"`
	Scrobbling ScrobblingConfig `toml:"scrobbling"`
}

// NavidromeConfig contains Navidrome server settings
type NavidromeConfig struct {
	ServerURL string `toml:"server_url"`
	Username  string `toml:"username"`
	Password  string `toml:"password"`
	Timeout   int    `toml:"timeout"` // in seconds
}

// AudioConfig contains audio playback settings
type AudioConfig struct {
	Device     string `toml:"device"`     // Audio device (auto-detect if empty)
	Volume     int    `toml:"volume"`     // Default volume (0-100)
	BufferSize int    `toml:"buffer_size"` // Buffer size for streaming
}

// UIConfig contains user interface settings
type UIConfig struct {
	Theme          string            `toml:"theme"`
	ShowAlbumArt   bool              `toml:"show_album_art"`
	HomeAlbumCount int               `toml:"home_album_count"` // Albums shown per section on home tab
	Keybindings    map[string]string `toml:"keybindings"`
}

// ScrobblingConfig contains scrobbling service settings
type ScrobblingConfig struct {
    // Method selects how scrobbling is performed: "auto", "server", "client", or "disabled"
    Method       string             `toml:"method"`
    LastFM       LastFMConfig       `toml:"lastfm"`
    ListenBrainz ListenBrainzConfig `toml:"listenbrainz"`
}

// LastFMConfig contains Last.fm scrobbling settings
type LastFMConfig struct {
	Enabled  bool   `toml:"enabled"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	APIKey   string `toml:"api_key"`
	Secret   string `toml:"secret"`
}

// ListenBrainzConfig contains ListenBrainz scrobbling settings
type ListenBrainzConfig struct {
	Enabled bool   `toml:"enabled"`
	Token   string `toml:"token"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
    return &Config{
		Navidrome: NavidromeConfig{
			ServerURL: "",
			Username:  "",
			Password:  "",
			Timeout:   30,
		},
		Audio: AudioConfig{
			Device:     "", // Auto-detect
			Volume:     100,
			BufferSize: 4096,
		},
		UI: UIConfig{
			Theme:          "dark",
			ShowAlbumArt:   false, // ASCII art not implemented yet
			HomeAlbumCount: 8,
			Keybindings: map[string]string{
				"quit":       "ctrl+c,q",
				"next_tab":   "tab",
				"prev_tab":   "shift+tab",
				"play_pause": "space",
				"next_track": "alt+right",
				"prev_track": "alt+left",
				"volume_up":  "shift+up",
				"volume_down": "shift+down",
				"seek_forward": "right",
				"seek_backward": "left",
				"toggle_shuffle": "alt+s",
				"stop": "ctrl+s",
			},
		},
        Scrobbling: ScrobblingConfig{
            Method: "auto",
            LastFM: LastFMConfig{
                Enabled:  false,
                Username: "",
                Password: "",
                APIKey:   "", // Users need to get their own API key
                Secret:   "", // Users need to get their own secret
            },
            ListenBrainz: ListenBrainzConfig{
                Enabled: false,
                Token:   "",
            },
        },
    }
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	
	navitoneDir := filepath.Join(configDir, "navitone-cli")
	if err := os.MkdirAll(navitoneDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(navitoneDir, "config.toml"), nil
}

// Load loads configuration from file, creating default if it doesn't exist
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	
	config := DefaultConfig()
	
	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := Save(config); err != nil {
			return nil, err
		}
		return config, nil
	}
	
	// Load existing config
	_, err = toml.DecodeFile(configPath, config)
	if err != nil {
		return nil, err
	}
	
	return config, nil
}

// Save saves configuration to file
func Save(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Basic validation
	if c.Navidrome.ServerURL == "" {
		return &ValidationError{Field: "navidrome.server_url", Message: "Server URL is required"}
	}
	
	if c.Navidrome.Username == "" {
		return &ValidationError{Field: "navidrome.username", Message: "Username is required"}
	}
	
	if c.Audio.Volume < 0 || c.Audio.Volume > 100 {
		return &ValidationError{Field: "audio.volume", Message: "Volume must be between 0 and 100"}
	}
	
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
