package models

import "time"

// Tab represents different views in the application
type Tab int

const (
	HomeTab Tab = iota
	AlbumsTab
	ArtistsTab
	TracksTab
	PlaylistsTab
	QueueTab
	ConfigTab
)

// String returns the string representation of a tab
func (t Tab) String() string {
	switch t {
	case HomeTab:
		return "Home"
	case AlbumsTab:
		return "Albums"
	case ArtistsTab:
		return "Artists"
	case TracksTab:
		return "Tracks"
	case PlaylistsTab:
		return "Playlists"
	case QueueTab:
		return "Queue"
	case ConfigTab:
		return "Config"
	default:
		return "Unknown"
	}
}

// Album represents a music album
type Album struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Artist      string    `json:"artist"`
	ArtistID    string    `json:"artistId"`
	Year        int       `json:"year"`
	Genre       string    `json:"genre"`
	Duration    int       `json:"duration"`
	TrackCount  int       `json:"songCount"`
	CreatedAt   time.Time `json:"created"`
	CoverArt    string    `json:"coverArt,omitempty"`
}

// Artist represents a music artist
type Artist struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AlbumCount int    `json:"albumCount"`
	StarredAt  *time.Time `json:"starred,omitempty"`
}

// Track represents a music track
type Track struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	ArtistID string `json:"artistId"`
	Album    string `json:"album"`
	AlbumID  string `json:"albumId"`
	Genre    string `json:"genre"`
	Year     int    `json:"year"`
	Duration int    `json:"duration"`
	Track    int    `json:"track"`
	Disc     int    `json:"discNumber"`
	Size     int64  `json:"size"`
	Suffix   string `json:"suffix"`
	BitRate  int    `json:"bitRate"`
	Path     string `json:"path"`
}

// Playlist represents a user playlist
type Playlist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   string    `json:"comment"`
	Public    bool      `json:"public"`
	SongCount int       `json:"songCount"`
	Duration  int       `json:"duration"`
	CreatedAt time.Time `json:"created"`
	ChangedAt time.Time `json:"changed"`
}

// AppState represents the current state of the application
type AppState struct {
	CurrentTab    Tab
	IsPlaying     bool
	CurrentTrack  *Track
	Queue         []Track
	Volume        int
	Position      time.Duration
	ShowHelp      bool
	ConfigForm    *ConfigFormState
	
	// Content state
	Albums        []Album
	Artists       []Artist
	Tracks        []Track
	Playlists     []Playlist
	
	// UI state
	LoadingAlbums  bool
	LoadingArtists bool
	LoadingTracks  bool
	LoadingError   string
	
	// Selection state
	SelectedAlbumIndex int
	SelectedArtistIndex int
	SelectedTrackIndex int
	SelectedQueueIndex int
}