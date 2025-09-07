package mpv

import (
	"fmt"
	"time"
)

// CommandWrapper provides high-level command methods for MPV IPC
type CommandWrapper struct {
	ipc *IPCClient
}

// NewCommandWrapper creates a new command wrapper
func NewCommandWrapper(ipc *IPCClient) *CommandWrapper {
	return &CommandWrapper{
		ipc: ipc,
	}
}

// Playback Commands

// LoadFile loads a file for playback
func (c *CommandWrapper) LoadFile(url string, mode string) error {
	if mode == "" {
		mode = "replace" // Default mode
	}
	_, err := c.ipc.SendCommand("loadfile", url, mode)
	return err
}

// Play starts or resumes playback
func (c *CommandWrapper) Play() error {
	return c.SetProperty("pause", false)
}

// Pause pauses playback
func (c *CommandWrapper) Pause() error {
	return c.SetProperty("pause", true)
}

// Stop stops playback
func (c *CommandWrapper) Stop() error {
	_, err := c.ipc.SendCommand("stop")
	return err
}

// TogglePause toggles pause state
func (c *CommandWrapper) TogglePause() error {
	_, err := c.ipc.SendCommand("cycle", "pause")
	return err
}

// Seeking Commands

// SeekAbsolute seeks to an absolute position in seconds
func (c *CommandWrapper) SeekAbsolute(seconds float64) error {
	_, err := c.ipc.SendCommand("seek", seconds, "absolute")
	return err
}

// SeekRelative seeks relative to current position
func (c *CommandWrapper) SeekRelative(seconds float64) error {
	_, err := c.ipc.SendCommand("seek", seconds, "relative")
	return err
}

// SeekPercent seeks to a percentage of the file
func (c *CommandWrapper) SeekPercent(percent float64) error {
	_, err := c.ipc.SendCommand("seek", percent, "absolute-percent")
	return err
}

// Volume Commands

// SetVolume sets the playback volume (0-100)
func (c *CommandWrapper) SetVolume(volume float64) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	return c.SetProperty("volume", volume)
}

// GetVolume gets the current volume
func (c *CommandWrapper) GetVolume() (float64, error) {
	result, err := c.GetProperty("volume")
	if err != nil {
		return 0, err
	}
	
	if vol, ok := result.(float64); ok {
		return vol, nil
	}
	return 0, fmt.Errorf("invalid volume type")
}

// Property Commands

// SetProperty sets an MPV property
func (c *CommandWrapper) SetProperty(property string, value interface{}) error {
	_, err := c.ipc.SendCommand("set_property", property, value)
	return err
}

// GetProperty gets an MPV property
func (c *CommandWrapper) GetProperty(property string) (interface{}, error) {
	return c.ipc.SendCommand("get_property", property)
}

// ObserveProperty observes an MPV property for changes
func (c *CommandWrapper) ObserveProperty(id int, property string) error {
	_, err := c.ipc.SendCommand("observe_property", id, property)
	return err
}

// UnobserveProperty stops observing an MPV property
func (c *CommandWrapper) UnobserveProperty(id int) error {
	_, err := c.ipc.SendCommand("unobserve_property", id)
	return err
}

// State Query Commands

// IsPlaying returns whether MPV is currently playing
func (c *CommandWrapper) IsPlaying() (bool, error) {
	paused, err := c.GetProperty("pause")
	if err != nil {
		return false, err
	}
	
	if pausedBool, ok := paused.(bool); ok {
		return !pausedBool, nil
	}
	return false, fmt.Errorf("invalid pause property type")
}

// GetPosition returns current playback position in seconds
func (c *CommandWrapper) GetPosition() (time.Duration, error) {
	result, err := c.GetProperty("playback-time")
	if err != nil {
		return 0, err
	}
	
	if pos, ok := result.(float64); ok {
		return time.Duration(pos * float64(time.Second)), nil
	}
	return 0, fmt.Errorf("invalid position type")
}

// GetDuration returns total duration in seconds
func (c *CommandWrapper) GetDuration() (time.Duration, error) {
	result, err := c.GetProperty("duration")
	if err != nil {
		return 0, err
	}
	
	if dur, ok := result.(float64); ok {
		return time.Duration(dur * float64(time.Second)), nil
	}
	return 0, fmt.Errorf("invalid duration type")
}

// GetFilename returns the currently loaded filename
func (c *CommandWrapper) GetFilename() (string, error) {
	result, err := c.GetProperty("filename")
	if err != nil {
		return "", err
	}
	
	if filename, ok := result.(string); ok {
		return filename, nil
	}
	return "", fmt.Errorf("invalid filename type")
}

// GetPath returns the currently loaded file path
func (c *CommandWrapper) GetPath() (string, error) {
	result, err := c.GetProperty("path")
	if err != nil {
		return "", err
	}
	
	if path, ok := result.(string); ok {
		return path, nil
	}
	return "", fmt.Errorf("invalid path type")
}

// Audio Commands

// SetAudioDevice sets the audio output device
func (c *CommandWrapper) SetAudioDevice(device string) error {
	return c.SetProperty("audio-device", device)
}

// GetAudioDevice gets the current audio output device
func (c *CommandWrapper) GetAudioDevice() (string, error) {
	result, err := c.GetProperty("audio-device")
	if err != nil {
		return "", err
	}
	
	if device, ok := result.(string); ok {
		return device, nil
	}
	return "", fmt.Errorf("invalid audio device type")
}

// SetReplayGain sets replay gain mode
func (c *CommandWrapper) SetReplayGain(mode string) error {
	// Valid modes: "no", "track", "album"
	return c.SetProperty("replaygain", mode)
}

// Playlist Commands (for future queue management)

// PlaylistNext plays next item in playlist
func (c *CommandWrapper) PlaylistNext() error {
	_, err := c.ipc.SendCommand("playlist-next")
	return err
}

// PlaylistPrev plays previous item in playlist
func (c *CommandWrapper) PlaylistPrev() error {
	_, err := c.ipc.SendCommand("playlist-prev")
	return err
}

// PlaylistClear clears the playlist
func (c *CommandWrapper) PlaylistClear() error {
	_, err := c.ipc.SendCommand("playlist-clear")
	return err
}

// PlaylistAppend adds a file to the playlist
func (c *CommandWrapper) PlaylistAppend(url string) error {
	_, err := c.ipc.SendCommand("loadfile", url, "append")
	return err
}

// PlaylistRemove removes an item from the playlist
func (c *CommandWrapper) PlaylistRemove(index int) error {
	_, err := c.ipc.SendCommand("playlist-remove", index)
	return err
}

// GetPlaylistCount gets the number of items in the playlist
func (c *CommandWrapper) GetPlaylistCount() (int, error) {
	result, err := c.GetProperty("playlist-count")
	if err != nil {
		return 0, err
	}
	
	if count, ok := result.(float64); ok {
		return int(count), nil
	}
	return 0, fmt.Errorf("invalid playlist count type")
}

// GetPlaylistPos gets the current playlist position
func (c *CommandWrapper) GetPlaylistPos() (int, error) {
	result, err := c.GetProperty("playlist-pos")
	if err != nil {
		return -1, err
	}
	
	if pos, ok := result.(float64); ok {
		return int(pos), nil
	}
	return -1, fmt.Errorf("invalid playlist position type")
}

// SetPlaylistPos sets the current playlist position
func (c *CommandWrapper) SetPlaylistPos(index int) error {
	return c.SetProperty("playlist-pos", index)
}

// Advanced Commands

// Screenshot takes a screenshot (useful for debugging)
func (c *CommandWrapper) Screenshot() error {
	_, err := c.ipc.SendCommand("screenshot")
	return err
}

// GetVersion gets MPV version
func (c *CommandWrapper) GetVersion() (string, error) {
	result, err := c.GetProperty("mpv-version")
	if err != nil {
		return "", err
	}
	
	if version, ok := result.(string); ok {
		return version, nil
	}
	return "", fmt.Errorf("invalid version type")
}

// Quit quits MPV
func (c *CommandWrapper) Quit() error {
	return c.ipc.SendCommandAsync("quit")
}

// QuitWatchLater quits MPV and saves the current position
func (c *CommandWrapper) QuitWatchLater() error {
	return c.ipc.SendCommandAsync("quit-watch-later")
}