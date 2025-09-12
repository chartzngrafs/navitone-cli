package scrobbling

// ScrobbleTrack represents a track for scrobbling
type ScrobbleTrack struct {
	Artist      string // Required
	Title       string // Required
	Album       string // Optional
	Duration    int    // Duration in seconds
	TrackNumber int    // Track number on album
	Timestamp   int64  // Unix timestamp when track was played
	MBID        string // MusicBrainz ID (optional)
}

// ScrobblingMethod selects how scrobbling should be performed
type ScrobblingMethod string

const (
    MethodAuto     ScrobblingMethod = "auto"
    MethodServer   ScrobblingMethod = "server"
    MethodClient   ScrobblingMethod = "client"
    MethodDisabled ScrobblingMethod = "disabled"
)

// UserInfo represents user information from Last.fm
type UserInfo struct {
    Name         string `json:"name"`
	RealName     string `json:"realname"`
	Country      string `json:"country"`
	PlayCount    int    `json:"playcount,string"`
	Registered   struct {
		UnixTime string `json:"unixtime"`
	} `json:"registered"`
}

// ScrobbleService defines the interface for scrobbling services
type ScrobbleService interface {
	// Scrobble submits a completed track play
	Scrobble(track ScrobbleTrack) error
	
	// UpdateNowPlaying updates the current playing status
	UpdateNowPlaying(track ScrobbleTrack) error
	
	// Name returns the service name
	Name() string
	
	// IsEnabled returns whether the service is enabled
	IsEnabled() bool
}

// ScrobbleResult represents the result of a scrobble operation
type ScrobbleResult struct {
	Service   string
	Success   bool
	Error     error
	Track     ScrobbleTrack
	Timestamp int64
}

// QueuedScrobble represents a scrobble that failed and is queued for retry
type QueuedScrobble struct {
	Track     ScrobbleTrack
	Service   string
	Attempts  int
	LastTry   int64
	MaxRetries int
}
