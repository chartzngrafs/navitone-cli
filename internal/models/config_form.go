package models

import (
	"fmt"
	"navitone-cli/internal/config"
)

// ConfigFormField represents a form field in the config tab
type ConfigFormField int

const (
	ServerURLField ConfigFormField = iota
	UsernameField
	PasswordField
	LastFMEnabledField
	LastFMUsernameField
	LastFMPasswordField
	ListenBrainzEnabledField
	ListenBrainzTokenField
	VolumeField
	AudioDeviceField
	BufferSizeField
)

// ConfigFormState represents the state of the configuration form
type ConfigFormState struct {
	Config          *config.Config
	ActiveField     ConfigFormField
	EditMode        bool
	CurrentInput    string
	ValidationError string
	TestingConnection bool
	ConnectionStatus  string
}

// NewConfigFormState creates a new config form state
func NewConfigFormState(cfg *config.Config) *ConfigFormState {
	return &ConfigFormState{
		Config:      cfg,
		ActiveField: ServerURLField,
		EditMode:    false,
	}
}

// GetFieldValue returns the current value for a form field
func (cfs *ConfigFormState) GetFieldValue(field ConfigFormField) string {
	switch field {
	case ServerURLField:
		return cfs.Config.Navidrome.ServerURL
	case UsernameField:
		return cfs.Config.Navidrome.Username
	case PasswordField:
		if cfs.Config.Navidrome.Password == "" {
			return ""
		}
		return "••••••••" // Masked password
	case LastFMUsernameField:
		return cfs.Config.Scrobbling.LastFM.Username
	case LastFMPasswordField:
		if cfs.Config.Scrobbling.LastFM.Password == "" {
			return ""
		}
		return "••••••••"
	case ListenBrainzTokenField:
		if cfs.Config.Scrobbling.ListenBrainz.Token == "" {
			return ""
		}
		return cfs.Config.Scrobbling.ListenBrainz.Token[:min(8, len(cfs.Config.Scrobbling.ListenBrainz.Token))] + "..."
	case VolumeField:
		return fmt.Sprintf("%d%%", cfs.Config.Audio.Volume)
	case AudioDeviceField:
		if cfs.Config.Audio.Device == "" {
			return "Auto-detect"
		}
		return cfs.Config.Audio.Device
	case BufferSizeField:
		return fmt.Sprintf("%d", cfs.Config.Audio.BufferSize)
	default:
		return ""
	}
}

// GetFieldLabel returns the label for a form field
func (cfs *ConfigFormState) GetFieldLabel(field ConfigFormField) string {
	switch field {
	case ServerURLField:
		return "Server URL"
	case UsernameField:
		return "Username"
	case PasswordField:
		return "Password"
	case LastFMUsernameField:
		return "Last.fm Username"
	case LastFMPasswordField:
		return "Last.fm Password"
	case ListenBrainzTokenField:
		return "ListenBrainz Token"
	case VolumeField:
		return "Volume"
	case AudioDeviceField:
		return "Audio Device"
	case BufferSizeField:
		return "Buffer Size"
	default:
		return ""
	}
}

// IsCheckboxField returns true if the field is a checkbox
func (cfs *ConfigFormState) IsCheckboxField(field ConfigFormField) bool {
	return field == LastFMEnabledField || field == ListenBrainzEnabledField
}

// GetCheckboxValue returns the checkbox value for boolean fields
func (cfs *ConfigFormState) GetCheckboxValue(field ConfigFormField) bool {
	switch field {
	case LastFMEnabledField:
		return cfs.Config.Scrobbling.LastFM.Enabled
	case ListenBrainzEnabledField:
		return cfs.Config.Scrobbling.ListenBrainz.Enabled
	default:
		return false
	}
}

// SetFieldValue updates a field value
func (cfs *ConfigFormState) SetFieldValue(field ConfigFormField, value string) {
	switch field {
	case ServerURLField:
		cfs.Config.Navidrome.ServerURL = value
	case UsernameField:
		cfs.Config.Navidrome.Username = value
	case PasswordField:
		cfs.Config.Navidrome.Password = value
	case LastFMUsernameField:
		cfs.Config.Scrobbling.LastFM.Username = value
	case LastFMPasswordField:
		cfs.Config.Scrobbling.LastFM.Password = value
	case ListenBrainzTokenField:
		cfs.Config.Scrobbling.ListenBrainz.Token = value
	case AudioDeviceField:
		cfs.Config.Audio.Device = value
	}
}

// ToggleCheckbox toggles a boolean field
func (cfs *ConfigFormState) ToggleCheckbox(field ConfigFormField) {
	switch field {
	case LastFMEnabledField:
		cfs.Config.Scrobbling.LastFM.Enabled = !cfs.Config.Scrobbling.LastFM.Enabled
	case ListenBrainzEnabledField:
		cfs.Config.Scrobbling.ListenBrainz.Enabled = !cfs.Config.Scrobbling.ListenBrainz.Enabled
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}