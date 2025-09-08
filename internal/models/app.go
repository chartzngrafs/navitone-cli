package models

import (
	"fmt"
	"time"
)

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

// SearchResults represents organized search results
type SearchResults struct {
	Artists []Artist
	Albums  []Album
	Tracks  []Track
}

// AppState represents the current state of the application
type AppState struct {
	CurrentTab    Tab
	IsPlaying     bool
	CurrentTrack  *Track
	Queue         []Track
	Volume        int
	Position      time.Duration
	IsShuffleMode bool
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
	
	// Modal state
	ShowAlbumModal      bool
	ShowArtistModal     bool
	ShowSearchModal     bool
	SelectedAlbum       *Album
	SelectedArtist      *Artist
	AlbumTracks         []Track
	ArtistAlbums        []Album
	SelectedModalIndex  int
	LoadingModalContent bool
	
	// Search state
	SearchQuery         string
	SearchResults       SearchResults
	LoadingSearchResults bool
	SelectedSearchIndex int
	SearchArtistsOffset int
	SearchAlbumsOffset  int
	SearchTracksOffset  int
	
	// Log state (for contained event logging)
	LogMessages []string
}

// AddLogMessage adds a log message to the log buffer, keeping only the latest messages
func (a *AppState) AddLogMessage(message string) {
	// Add timestamp prefix for better user experience
	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	
	a.LogMessages = append(a.LogMessages, formattedMessage)
	
	// Keep only the latest 10 messages (only 2 are shown, but we keep more for scroll-back potential)
	if len(a.LogMessages) > 10 {
		a.LogMessages = a.LogMessages[len(a.LogMessages)-10:]
	}
}