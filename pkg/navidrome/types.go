package navidrome

import "time"

// SubsonicError represents an error response from the Subsonic API
type SubsonicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BaseResponse contains common response fields
type BaseResponse struct {
	Status string         `json:"status"`
	Error  *SubsonicError `json:"error,omitempty"`
}

// Album represents an album from Navidrome
type Album struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Artist      string    `json:"artist"`
	ArtistID    string    `json:"artistId"`
	CoverArt    string    `json:"coverArt,omitempty"`
	SongCount   int       `json:"songCount"`
	Duration    int       `json:"duration"`
	PlayCount   int       `json:"playCount"`
	Created     time.Time `json:"created"`
	Year        int       `json:"year,omitempty"`
	Genre       string    `json:"genre,omitempty"`
}

// Artist represents an artist from Navidrome
type Artist struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CoverArt   string `json:"coverArt,omitempty"`
	AlbumCount int    `json:"albumCount"`
	Starred    *time.Time `json:"starred,omitempty"`
}

// ArtistWithAlbums represents an artist with nested albums
type ArtistWithAlbums struct {
	Artist
	Album []Album `json:"album,omitempty"`
}

// Song represents a song/track from Navidrome
type Song struct {
	ID          string    `json:"id"`
	Parent      string    `json:"parent"`
	Title       string    `json:"title"`
	Album       string    `json:"album"`
	Artist      string    `json:"artist"`
	Track       int       `json:"track,omitempty"`
	Year        int       `json:"year,omitempty"`
	Genre       string    `json:"genre,omitempty"`
	CoverArt    string    `json:"coverArt,omitempty"`
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	Suffix      string    `json:"suffix"`
	Duration    int       `json:"duration"`
	BitRate     int       `json:"bitRate,omitempty"`
	Path        string    `json:"path"`
	IsDir       bool      `json:"isDir"`
	AlbumID     string    `json:"albumId"`
	ArtistID    string    `json:"artistId"`
	Type        string    `json:"type"`
	Created     time.Time `json:"created"`
	PlayCount   int       `json:"playCount,omitempty"`
	DiscNumber  int       `json:"discNumber,omitempty"`
	Starred     *time.Time `json:"starred,omitempty"`
}

// Playlist represents a playlist from Navidrome
type Playlist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   string    `json:"comment,omitempty"`
	Owner     string    `json:"owner"`
	Public    bool      `json:"public"`
	SongCount int       `json:"songCount"`
	Duration  int       `json:"duration"`
	Created   time.Time `json:"created"`
	Changed   time.Time `json:"changed"`
	CoverArt  string    `json:"coverArt,omitempty"`
	Entry     []Song    `json:"entry,omitempty"`
}

// AlbumList contains a list of albums
type AlbumList struct {
	Album []Album `json:"album"`
}

// ArtistsList contains a list of artists organized by index
type ArtistsList struct {
	Index []struct {
		Name   string   `json:"name"`
		Artist []Artist `json:"artist"`
	} `json:"index"`
}

// SongsList contains a list of songs
type SongsList struct {
	Song []Song `json:"song"`
}

// PlaylistsList contains a list of playlists
type PlaylistsList struct {
	Playlist []Playlist `json:"playlist"`
}

// Response types for different API endpoints

// AlbumsResponse represents the response from getAlbumList2
type AlbumsResponse struct {
	SubsonicResponse struct {
		BaseResponse
		AlbumList2 AlbumList `json:"albumList2"`
	} `json:"subsonic-response"`
}

// ArtistsResponse represents the response from getArtists
type ArtistsResponse struct {
	SubsonicResponse struct {
		BaseResponse
		Artists ArtistsList `json:"artists"`
	} `json:"subsonic-response"`
}

// SongsResponse represents the response from getSongsByGenre or similar
type SongsResponse struct {
	SubsonicResponse struct {
		BaseResponse
		SongsByGenre SongsList `json:"songsByGenre"`
	} `json:"subsonic-response"`
}

// PlaylistsResponse represents the response from getPlaylists
type PlaylistsResponse struct {
	SubsonicResponse struct {
		BaseResponse
		Playlists PlaylistsList `json:"playlists"`
	} `json:"subsonic-response"`
}

// PlaylistResponse represents the response from getPlaylist
type PlaylistResponse struct {
	SubsonicResponse struct {
		BaseResponse
		Playlist Playlist `json:"playlist"`
	} `json:"subsonic-response"`
}

// SearchResult represents search results
type SearchResult struct {
	Artist   []Artist `json:"artist,omitempty"`
	Album    []Album  `json:"album,omitempty"`
	Song     []Song   `json:"song,omitempty"`
}

// SearchResponse represents the response from search3
type SearchResponse struct {
	SubsonicResponse struct {
		BaseResponse
		SearchResult3 SearchResult `json:"searchResult3"`
	} `json:"subsonic-response"`
}

// RandomSongsResponse represents the response from getRandomSongs
type RandomSongsResponse struct {
	SubsonicResponse struct {
		BaseResponse
		RandomSongs SongsList `json:"randomSongs"`
	} `json:"subsonic-response"`
}

// User represents a user from Navidrome
type User struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	ScrobblingEnabled    bool   `json:"scrobblingEnabled"`
	MaxBitRate           int    `json:"maxBitRate"`
	AdminRole            bool   `json:"adminRole"`
	SettingsRole         bool   `json:"settingsRole"`
	DownloadRole         bool   `json:"downloadRole"`
	UploadRole           bool   `json:"uploadRole"`
	PlaylistRole         bool   `json:"playlistRole"`
	CoverArtRole         bool   `json:"coverArtRole"`
	CommentRole          bool   `json:"commentRole"`
	PodcastRole          bool   `json:"podcastRole"`
	StreamRole           bool   `json:"streamRole"`
	JukeboxRole          bool   `json:"jukeboxRole"`
	ShareRole            bool   `json:"shareRole"`
	VideoConversionRole  bool   `json:"videoConversionRole"`
}

// UserResponse represents the response from getUser
type UserResponse struct {
    SubsonicResponse struct {
        BaseResponse
        User User `json:"user"`
    } `json:"subsonic-response"`
}

// ScrobblingCapabilities represents server-side scrobbling availability and user status
type ScrobblingCapabilities struct {
    ServerScrobblingAvailable bool
    UserScrobblingEnabled     bool
    SupportedServices         []string
}
