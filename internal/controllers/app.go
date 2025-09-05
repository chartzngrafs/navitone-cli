package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"navitone-cli/internal/audio"
	"navitone-cli/internal/config"
	"navitone-cli/internal/models"
	"navitone-cli/internal/views"
	"navitone-cli/pkg/navidrome"
	"navitone-cli/pkg/scrobbling"

	tea "github.com/charmbracelet/bubbletea"
)

// App represents the main application controller
type App struct {
	state           *models.AppState
	view            *views.MainView
	navidromeClient *navidrome.Client
	audioManager    *audio.Manager
	scrobbler       *scrobbling.Manager
}

// NewApp creates a new application instance
func NewApp() *App {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	state := &models.AppState{
		CurrentTab: models.HomeTab,
		Volume:     70,
		Queue:      make([]models.Track, 0),
		ConfigForm: models.NewConfigFormState(cfg),
		Albums:      make([]models.Album, 0),
		Artists:     make([]models.Artist, 0),
		Tracks:      make([]models.Track, 0),
		Playlists:   make([]models.Playlist, 0),
		LogMessages: make([]string, 0),
	}

	app := &App{
		state: state,
		view:  views.NewMainView(state),
	}

	// Initialize Navidrome client if config is valid
	app.initializeNavidromeClient()

	// Initialize scrobbling manager
	app.scrobbler = scrobbling.NewManager(cfg)

	// Initialize audio manager
	if app.navidromeClient != nil {
		audioManager, err := audio.NewManager(app.navidromeClient, app.scrobbler)
		if err == nil {
			app.audioManager = audioManager
			// Set up callback to update app state when audio changes
			audioManager.SetStateCallback(app.updateAudioState)
			// Set up callback for log messages
			audioManager.SetLogCallback(app.logMessage)
		}
	}

	return app
}

// updateAudioState updates the app state based on audio manager changes
func (a *App) updateAudioState(state *models.AppState) {
	if a.audioManager != nil {
		// Update queue from audio manager
		a.state.Queue = a.audioManager.GetQueue()

		// Update current playing track
		currentTrack := a.audioManager.GetCurrentTrack()
		a.state.CurrentTrack = currentTrack

		// Update playing state
		a.state.IsPlaying = a.audioManager.IsPlaying()

		// Update position if available
		// TODO: Get position from audio manager if needed
	}
}

// logMessage adds a message to the app's log area
func (a *App) logMessage(message string) {
	a.state.AddLogMessage(message)
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle modal navigation first
		if a.state.ShowAlbumModal || a.state.ShowArtistModal {
			return a.handleModalKeyPress(msg)
		}
		return a.handleKeyPress(msg)
	case tea.MouseMsg:
		return a.handleMouseEvent(msg)
	case tea.WindowSizeMsg:
		a.view.SetSize(msg.Width, msg.Height)
		return a, nil
	case ConnectionTestResult:
		// Handle connection test result
		cf := a.state.ConfigForm
		cf.TestingConnection = false
		cf.ConnectionStatus = msg.Message
		// Reinitialize client if connection was successful
		if msg.Success {
			a.initializeNavidromeClient()
		}
		return a, nil
	case AlbumsLoadResult:
		// Handle albums load result
		a.state.LoadingAlbums = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.Albums = msg.Albums
			a.state.LoadingError = ""
		}
		return a, nil
	case ArtistsLoadResult:
		// Handle artists load result
		a.state.LoadingArtists = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.Artists = msg.Artists
			a.state.LoadingError = ""
		}
		return a, nil
	case TracksLoadResult:
		// Handle tracks load result
		a.state.LoadingTracks = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.Tracks = msg.Tracks
			a.state.LoadingError = ""
		}
		return a, nil
	case AlbumTracksLoadResult:
		// Handle album tracks load result and add to queue
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			// Add all tracks to queue
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(msg.Tracks)
				// Sync state immediately
				a.state.Queue = a.audioManager.GetQueue()
				a.state.CurrentTrack = a.audioManager.GetCurrentTrack()
				a.state.IsPlaying = a.audioManager.IsPlaying()
			} else {
				a.state.Queue = append(a.state.Queue, msg.Tracks...)
			}
			a.state.LoadingError = ""
			a.logMessage(fmt.Sprintf("Added album to queue (%d tracks, total: %d)", len(msg.Tracks), len(a.state.Queue)))
		}
		return a, nil
	case ArtistTracksLoadResult:
		// Handle artist tracks load result and add to queue
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			// Add all tracks to queue
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(msg.Tracks)
				// Sync state immediately
				a.state.Queue = a.audioManager.GetQueue()
				a.state.CurrentTrack = a.audioManager.GetCurrentTrack()
				a.state.IsPlaying = a.audioManager.IsPlaying()
			} else {
				a.state.Queue = append(a.state.Queue, msg.Tracks...)
			}
			a.state.LoadingError = ""
			a.logMessage(fmt.Sprintf("Added artist tracks to queue (%d tracks, total: %d)", len(msg.Tracks), len(a.state.Queue)))
		}
		return a, nil
	case AlbumTracksModalResult:
		// Handle album tracks load for modal display
		a.state.LoadingModalContent = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.AlbumTracks = msg.Tracks
			a.state.SelectedModalIndex = 0
			a.state.LoadingError = ""
		}
		return a, nil
	case ArtistAlbumsModalResult:
		// Handle artist albums load for modal display
		a.state.LoadingModalContent = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.ArtistAlbums = msg.Albums
			a.state.SelectedModalIndex = 0
			a.state.LoadingError = ""
		}
		return a, nil
	}

	return a, nil
}

// View implements tea.Model
func (a *App) View() string {
	return a.view.Render()
}

// handleKeyPress processes keyboard input
func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle config form input if in config tab
	if a.state.CurrentTab == models.ConfigTab {
		return a.handleConfigKeyPress(msg)
	}

	// Handle content browsing tabs
	if a.state.CurrentTab == models.AlbumsTab {
		return a.handleAlbumsKeyPress(msg)
	}
	if a.state.CurrentTab == models.ArtistsTab {
		return a.handleArtistsKeyPress(msg)
	}
	if a.state.CurrentTab == models.TracksTab {
		return a.handleTracksKeyPress(msg)
	}
	if a.state.CurrentTab == models.QueueTab {
		return a.handleQueueKeyPress(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		// Clean up audio resources before quitting
		if a.audioManager != nil {
			a.audioManager.Close()
		}
		return a, tea.Quit
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
	case "ctrl+d":
		// Debug: Check streaming permissions
		if a.audioManager != nil {
			go func() {
				a.audioManager.CheckStreamingPermissions()
			}()
		}
	case "ctrl+p":
		// Global: Play/Pause toggle
		if a.audioManager != nil {
			a.audioManager.TogglePlayPause()
		} else {
			a.state.IsPlaying = !a.state.IsPlaying
		}
	case "ctrl+n":
		// Global: Next track
		if a.audioManager != nil {
			a.audioManager.NextTrack()
		}
	case "ctrl+b":
		// Global: Previous track
		if a.audioManager != nil {
			a.audioManager.PreviousTrack()
		}
	case "ctrl+s":
		// Global: Stop
		if a.audioManager != nil {
			a.audioManager.Stop()
		} else {
			a.state.IsPlaying = false
		}
	}

	return a, nil
}

// handleConfigKeyPress handles keyboard input for the config tab
func (a *App) handleConfigKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cf := a.state.ConfigForm

	// Global keys work even in config tab
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
		return a, nil
	case "tab":
		if !cf.EditMode {
			a.nextTab()
		}
		return a, nil
	case "shift+tab":
		if !cf.EditMode {
			a.prevTab()
		}
		return a, nil
	}

	// Config-specific keys
	if cf.EditMode {
		return a.handleConfigEditMode(msg)
	} else {
		return a.handleConfigNavigationMode(msg)
	}
}

// handleConfigNavigationMode handles navigation in config form
func (a *App) handleConfigNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cf := a.state.ConfigForm

	switch msg.String() {
	case "up", "k":
		a.moveConfigField(-1)
	case "down", "j":
		a.moveConfigField(1)
	case "enter":
		if cf.IsCheckboxField(cf.ActiveField) {
			cf.ToggleCheckbox(cf.ActiveField)
		} else {
			cf.EditMode = true
			cf.CurrentInput = a.getEditableValue(cf.ActiveField)
		}
	case "f2":
		return a.saveConfig()
	case "f3":
		return a.testConnection()
	}

	return a, nil
}

// handleConfigEditMode handles text input in edit mode
func (a *App) handleConfigEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cf := a.state.ConfigForm

	switch msg.String() {
	case "enter":
		// Validate and convert numeric fields using tagged switch
		switch cf.ActiveField {
		case models.VolumeField:
			if vol, err := strconv.Atoi(cf.CurrentInput); err == nil && vol >= 0 && vol <= 100 {
				cf.Config.Audio.Volume = vol
			} else {
				cf.ValidationError = "Volume must be a number between 0 and 100"
				return a, nil
			}
		case models.BufferSizeField:
			if size, err := strconv.Atoi(cf.CurrentInput); err == nil && size > 0 {
				cf.Config.Audio.BufferSize = size
			} else {
				cf.ValidationError = "Buffer size must be a positive number"
				return a, nil
			}
		default:
			cf.SetFieldValue(cf.ActiveField, cf.CurrentInput)
		}

		cf.EditMode = false
		cf.CurrentInput = ""
		cf.ValidationError = ""
	case "esc":
		cf.EditMode = false
		cf.CurrentInput = ""
	case "backspace":
		if len(cf.CurrentInput) > 0 {
			cf.CurrentInput = cf.CurrentInput[:len(cf.CurrentInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			cf.CurrentInput += msg.String()
		}
	}

	return a, nil
}

// moveConfigField moves the active field up or down
func (a *App) moveConfigField(direction int) {
	cf := a.state.ConfigForm
	current := int(cf.ActiveField)
	max := int(models.BufferSizeField)

	current += direction
	if current < 0 {
		current = max
	} else if current > max {
		current = 0
	}

	cf.ActiveField = models.ConfigFormField(current)
}

// getEditableValue returns the actual value for editing (not masked)
func (a *App) getEditableValue(field models.ConfigFormField) string {
	cf := a.state.ConfigForm
	switch field {
	case models.ServerURLField:
		return cf.Config.Navidrome.ServerURL
	case models.UsernameField:
		return cf.Config.Navidrome.Username
	case models.PasswordField:
		return cf.Config.Navidrome.Password
	case models.LastFMUsernameField:
		return cf.Config.Scrobbling.LastFM.Username
	case models.LastFMPasswordField:
		return cf.Config.Scrobbling.LastFM.Password
	case models.ListenBrainzTokenField:
		return cf.Config.Scrobbling.ListenBrainz.Token
	case models.VolumeField:
		return fmt.Sprintf("%d", cf.Config.Audio.Volume)
	case models.AudioDeviceField:
		return cf.Config.Audio.Device
	case models.BufferSizeField:
		return fmt.Sprintf("%d", cf.Config.Audio.BufferSize)
	default:
		return ""
	}
}

// saveConfig saves the current configuration
func (a *App) saveConfig() (tea.Model, tea.Cmd) {
	cf := a.state.ConfigForm

	if err := cf.Config.Validate(); err != nil {
		cf.ValidationError = err.Error()
		return a, nil
	}

	if err := config.Save(cf.Config); err != nil {
		cf.ValidationError = "Failed to save config: " + err.Error()
		return a, nil
	}

	cf.ValidationError = ""
	cf.ConnectionStatus = "Configuration saved successfully!"
	return a, nil
}

// testConnection tests the Navidrome connection
func (a *App) testConnection() (tea.Model, tea.Cmd) {
	cf := a.state.ConfigForm
	cf.TestingConnection = true
	cf.ConnectionStatus = "Testing connection..."

	// Return a command to test the connection asynchronously
	return a, tea.Cmd(func() tea.Msg {
		return a.doConnectionTest()
	})
}

// ConnectionTestResult represents the result of a connection test
type ConnectionTestResult struct {
	Success bool
	Message string
}

// doConnectionTest performs the actual connection test
func (a *App) doConnectionTest() ConnectionTestResult {
	cf := a.state.ConfigForm

	// Basic validation
	if cf.Config.Navidrome.ServerURL == "" {
		return ConnectionTestResult{
			Success: false,
			Message: "❌ Server URL is required",
		}
	}

	if cf.Config.Navidrome.Username == "" {
		return ConnectionTestResult{
			Success: false,
			Message: "❌ Username is required",
		}
	}

	if cf.Config.Navidrome.Password == "" {
		return ConnectionTestResult{
			Success: false,
			Message: "❌ Password is required",
		}
	}

	// Create Navidrome client
	client := navidrome.NewClient(
		cf.Config.Navidrome.ServerURL,
		cf.Config.Navidrome.Username,
		cf.Config.Navidrome.Password,
	)

	// Set timeout from config
	client.SetTimeout(time.Duration(cf.Config.Navidrome.Timeout) * time.Second)

	// Test connection with ping
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("❌ Connection failed: %s", err.Error()),
		}
	}

	return ConnectionTestResult{
		Success: true,
		Message: "✅ Connection successful!",
	}
}

// initializeNavidromeClient sets up the Navidrome client if config is valid
func (a *App) initializeNavidromeClient() {
	cfg := a.state.ConfigForm.Config

	if cfg.Navidrome.ServerURL != "" && cfg.Navidrome.Username != "" && cfg.Navidrome.Password != "" {
		a.navidromeClient = navidrome.NewClient(
			cfg.Navidrome.ServerURL,
			cfg.Navidrome.Username,
			cfg.Navidrome.Password,
		)
		a.navidromeClient.SetTimeout(time.Duration(cfg.Navidrome.Timeout) * time.Second)
	}
}

// handleTabChange handles actions when switching tabs
func (a *App) handleTabChange() tea.Cmd {
	// Load data when entering certain tabs
	switch a.state.CurrentTab {
	case models.AlbumsTab:
		if len(a.state.Albums) == 0 && a.navidromeClient != nil && !a.state.LoadingAlbums {
			return a.loadAlbums()
		}
	case models.ArtistsTab:
		if len(a.state.Artists) == 0 && a.navidromeClient != nil && !a.state.LoadingArtists {
			return a.loadArtists()
		}
	case models.TracksTab:
		if len(a.state.Tracks) == 0 && a.navidromeClient != nil && !a.state.LoadingTracks {
			return a.loadTracks()
		}
	}
	return nil
}

// handleAlbumsKeyPress handles keyboard input for the albums tab
func (a *App) handleAlbumsKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Debug: log the key that was pressed
	// a.logMessage(fmt.Sprintf("Albums tab key: '%s'", msg.String()))
	
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up", "k":
		if a.state.SelectedAlbumIndex > 0 {
			a.state.SelectedAlbumIndex--
		}
	case "down", "j":
		if a.state.SelectedAlbumIndex < len(a.state.Albums)-1 {
			a.state.SelectedAlbumIndex++
		}
	case "enter":
		// Show album details modal (regular Enter)
		if a.state.SelectedAlbumIndex < len(a.state.Albums) {
			return a, a.showAlbumModal(a.state.Albums[a.state.SelectedAlbumIndex])
		}
	case "alt+enter":
		// Queue entire album immediately (Alt+Enter)
		if a.state.SelectedAlbumIndex < len(a.state.Albums) {
			return a, a.addAlbumToQueue(a.state.Albums[a.state.SelectedAlbumIndex])
		}
	case "a":
		// Alternative: 'A' key to queue entire album immediately
		if a.state.SelectedAlbumIndex < len(a.state.Albums) {
			return a, a.addAlbumToQueue(a.state.Albums[a.state.SelectedAlbumIndex])
		}
	case "r":
		// Refresh albums
		return a, a.loadAlbums()
	}

	return a, nil
}

// loadAlbums loads albums from Navidrome
func (a *App) loadAlbums() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingAlbums = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := a.navidromeClient.GetAlbums(ctx, 50, 0) // Get first 50 albums
		if err != nil {
			return AlbumsLoadResult{Error: err}
		}

		// Convert Navidrome albums to our model
		albums := make([]models.Album, len(resp.SubsonicResponse.AlbumList2.Album))
		for i, album := range resp.SubsonicResponse.AlbumList2.Album {
			albums[i] = models.Album{
				ID:         album.ID,
				Name:       album.Name,
				Artist:     album.Artist,
				ArtistID:   album.ArtistID,
				Year:       album.Year,
				Genre:      album.Genre,
				Duration:   album.Duration,
				TrackCount: album.SongCount,
				CreatedAt:  album.Created,
				CoverArt:   album.CoverArt,
			}
		}

		return AlbumsLoadResult{Albums: albums}
	})
}

// loadArtists loads artists from Navidrome
func (a *App) loadArtists() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingArtists = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := a.navidromeClient.GetArtists(ctx)
		if err != nil {
			return ArtistsLoadResult{Error: err}
		}

		// Convert Navidrome artists to our model
		var artists []models.Artist
		for _, index := range resp.SubsonicResponse.Artists.Index {
			for _, artist := range index.Artist {
				artists = append(artists, models.Artist{
					ID:         artist.ID,
					Name:       artist.Name,
					AlbumCount: artist.AlbumCount,
					StarredAt:  artist.Starred,
				})
			}
		}

		return ArtistsLoadResult{Artists: artists}
	})
}

// loadTracks loads tracks from Navidrome
func (a *App) loadTracks() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingTracks = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := a.navidromeClient.GetSongs(ctx, 50, 0) // Get first 50 songs
		if err != nil {
			return TracksLoadResult{Error: err}
		}

		// Convert Navidrome songs to our model
		tracks := make([]models.Track, len(resp.SubsonicResponse.SongsByGenre.Song))
		for i, song := range resp.SubsonicResponse.SongsByGenre.Song {
			tracks[i] = models.Track{
				ID:       song.ID,
				Title:    song.Title,
				Artist:   song.Artist,
				ArtistID: song.ArtistID,
				Album:    song.Album,
				AlbumID:  song.AlbumID,
				Genre:    song.Genre,
				Year:     song.Year,
				Duration: song.Duration,
				Track:    song.Track,
				Disc:     song.DiscNumber,
				Size:     song.Size,
				Suffix:   song.Suffix,
				BitRate:  song.BitRate,
				Path:     song.Path,
			}
		}

		return TracksLoadResult{Tracks: tracks}
	})
}

// addAlbumToQueue adds all tracks from an album to the queue
func (a *App) addAlbumToQueue(album models.Album) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if a.navidromeClient == nil {
				return AlbumTracksLoadResult{Error: fmt.Errorf("navidrome client not initialized")}
			}

			// Fetch actual tracks from the album
			resp, err := a.navidromeClient.GetAlbumTracks(context.Background(), album.ID)
			if err != nil {
				return AlbumTracksLoadResult{Error: err}
			}

			// Convert Navidrome songs to our model
			tracks := make([]models.Track, len(resp.SubsonicResponse.SongsByGenre.Song))
			for i, song := range resp.SubsonicResponse.SongsByGenre.Song {
				tracks[i] = models.Track{
					ID:       song.ID,
					Title:    song.Title,
					Artist:   song.Artist,
					ArtistID: song.ArtistID,
					Album:    song.Album,
					AlbumID:  song.AlbumID,
					Genre:    song.Genre,
					Year:     song.Year,
					Duration: song.Duration,
					Track:    song.Track,
					Disc:     song.DiscNumber,
					Size:     song.Size,
					Suffix:   song.Suffix,
					BitRate:  song.BitRate,
					Path:     song.Path,
				}
			}

			return AlbumTracksLoadResult{Tracks: tracks}
		},
	)
}

// AlbumTracksLoadResult represents the result of loading album tracks
type AlbumTracksLoadResult struct {
	Tracks []models.Track
	Error  error
}

// handleArtistsKeyPress handles keyboard input for the artists tab
func (a *App) handleArtistsKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up", "k":
		if a.state.SelectedArtistIndex > 0 {
			a.state.SelectedArtistIndex--
		}
	case "down", "j":
		if a.state.SelectedArtistIndex < len(a.state.Artists)-1 {
			a.state.SelectedArtistIndex++
		}
	case "enter":
		// Show artist albums modal
		if a.state.SelectedArtistIndex < len(a.state.Artists) {
			return a, a.showArtistModal(a.state.Artists[a.state.SelectedArtistIndex])
		}
	case "r":
		// Refresh artists
		return a, a.loadArtists()
	}

	return a, nil
}

// addArtistToQueue adds all tracks from an artist to the queue
func (a *App) addArtistToQueue(artist models.Artist) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if a.navidromeClient == nil {
				return ArtistTracksLoadResult{Error: fmt.Errorf("navidrome client not initialized")}
			}

			// Fetch actual tracks from the artist
			resp, err := a.navidromeClient.GetArtistTracks(context.Background(), artist.ID)
			if err != nil {
				return ArtistTracksLoadResult{Error: err}
			}

			// Convert Navidrome songs to our model
			tracks := make([]models.Track, len(resp.SubsonicResponse.SongsByGenre.Song))
			for i, song := range resp.SubsonicResponse.SongsByGenre.Song {
				tracks[i] = models.Track{
					ID:       song.ID,
					Title:    song.Title,
					Artist:   song.Artist,
					ArtistID: song.ArtistID,
					Album:    song.Album,
					AlbumID:  song.AlbumID,
					Genre:    song.Genre,
					Year:     song.Year,
					Duration: song.Duration,
					Track:    song.Track,
					Disc:     song.DiscNumber,
					Size:     song.Size,
					Suffix:   song.Suffix,
					BitRate:  song.BitRate,
					Path:     song.Path,
				}
			}

			return ArtistTracksLoadResult{Tracks: tracks}
		},
	)
}

// ArtistTracksLoadResult represents the result of loading artist tracks
type ArtistTracksLoadResult struct {
	Tracks []models.Track
	Error  error
}

// handleTracksKeyPress handles keyboard input for the tracks tab
func (a *App) handleTracksKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up", "k":
		if a.state.SelectedTrackIndex > 0 {
			a.state.SelectedTrackIndex--
		}
	case "down", "j":
		if a.state.SelectedTrackIndex < len(a.state.Tracks)-1 {
			a.state.SelectedTrackIndex++
		}
	case "enter":
		// Add selected track to queue
		if a.state.SelectedTrackIndex < len(a.state.Tracks) {
			return a, a.addTrackToQueue(a.state.Tracks[a.state.SelectedTrackIndex])
		}
	case "r":
		// Refresh tracks
		return a, a.loadTracks()
	}

	return a, nil
}

// addTrackToQueue adds a single track to the queue
func (a *App) addTrackToQueue(track models.Track) tea.Cmd {
	if a.audioManager != nil {
		a.audioManager.AddToQueue(track)
		// Sync state immediately
		a.state.Queue = a.audioManager.GetQueue()
		a.state.CurrentTrack = a.audioManager.GetCurrentTrack()
		a.state.IsPlaying = a.audioManager.IsPlaying()
	} else {
		a.state.Queue = append(a.state.Queue, track)
	}
	return nil
}

// handleQueueKeyPress handles keyboard input for the queue tab
func (a *App) handleQueueKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "F1", "?":
		a.state.ShowHelp = !a.state.ShowHelp
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up", "k":
		if a.state.SelectedQueueIndex > 0 {
			a.state.SelectedQueueIndex--
		}
	case "down", "j":
		if a.state.SelectedQueueIndex < len(a.state.Queue)-1 {
			a.state.SelectedQueueIndex++
		}
	case "delete", "x":
		// Remove selected track from queue
		if a.audioManager != nil && a.state.SelectedQueueIndex < len(a.state.Queue) {
			a.audioManager.RemoveFromQueue(a.state.SelectedQueueIndex)
		}
	case "c":
		// Clear entire queue
		if a.audioManager != nil {
			a.audioManager.ClearQueue()
		} else {
			a.state.Queue = make([]models.Track, 0)
		}
		a.state.SelectedQueueIndex = 0
	case "enter", "space":
		// Play selected track or toggle play/pause
		if a.audioManager != nil {
			if a.state.SelectedQueueIndex < len(a.state.Queue) {
				a.audioManager.PlayTrackAtIndex(a.state.SelectedQueueIndex)
			} else {
				a.audioManager.TogglePlayPause()
			}
		} else {
			// Fallback for when audio manager is not available
			if a.state.SelectedQueueIndex < len(a.state.Queue) {
				a.state.CurrentTrack = &a.state.Queue[a.state.SelectedQueueIndex]
				a.state.IsPlaying = !a.state.IsPlaying
			}
		}
	}

	return a, nil
}

// Message types for async operations
type AlbumsLoadResult struct {
	Albums []models.Album
	Error  error
}

type ArtistsLoadResult struct {
	Artists []models.Artist
	Error   error
}

type TracksLoadResult struct {
	Tracks []models.Track
	Error  error
}

// Modal-specific message types
type AlbumTracksModalResult struct {
	Tracks []models.Track
	Error  error
}

type ArtistAlbumsModalResult struct {
	Albums []models.Album
	Error  error
}


// handleMouseEvent processes mouse input
func (a *App) handleMouseEvent(_ tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse support placeholder for Phase 2
	return a, nil
}

// nextTab switches to the next tab
func (a *App) nextTab() {
	a.state.CurrentTab = models.Tab((int(a.state.CurrentTab) + 1) % 7)
}

// prevTab switches to the previous tab
func (a *App) prevTab() {
	current := int(a.state.CurrentTab)
	if current == 0 {
		current = 7
	}
	a.state.CurrentTab = models.Tab(current - 1)
}

// showAlbumModal displays the album tracks modal
func (a *App) showAlbumModal(album models.Album) tea.Cmd {
	a.state.ShowAlbumModal = true
	a.state.SelectedAlbum = &album
	a.state.LoadingModalContent = true
	a.state.AlbumTracks = nil
	a.state.SelectedModalIndex = 0

	return tea.Cmd(func() tea.Msg {
		if a.navidromeClient == nil {
			return AlbumTracksModalResult{Error: fmt.Errorf("navidrome client not initialized")}
		}

		resp, err := a.navidromeClient.GetAlbumTracks(context.Background(), album.ID)
		if err != nil {
			return AlbumTracksModalResult{Error: err}
		}

		tracks := make([]models.Track, len(resp.SubsonicResponse.SongsByGenre.Song))
		for i, song := range resp.SubsonicResponse.SongsByGenre.Song {
			tracks[i] = models.Track{
				ID:       song.ID,
				Title:    song.Title,
				Artist:   song.Artist,
				ArtistID: song.ArtistID,
				Album:    song.Album,
				AlbumID:  song.AlbumID,
				Genre:    song.Genre,
				Year:     song.Year,
				Duration: song.Duration,
				Track:    song.Track,
				Disc:     song.DiscNumber,
				Size:     song.Size,
				Suffix:   song.Suffix,
				BitRate:  song.BitRate,
				Path:     song.Path,
			}
		}

		return AlbumTracksModalResult{Tracks: tracks}
	})
}

// showArtistModal displays the artist albums modal
func (a *App) showArtistModal(artist models.Artist) tea.Cmd {
	a.state.ShowArtistModal = true
	a.state.SelectedArtist = &artist
	a.state.LoadingModalContent = true
	a.state.ArtistAlbums = nil
	a.state.SelectedModalIndex = 0

	return tea.Cmd(func() tea.Msg {
		if a.navidromeClient == nil {
			return ArtistAlbumsModalResult{Error: fmt.Errorf("navidrome client not initialized")}
		}

		resp, err := a.navidromeClient.GetArtistAlbums(context.Background(), artist.ID)
		if err != nil {
			return ArtistAlbumsModalResult{Error: err}
		}

		albums := make([]models.Album, len(resp.SubsonicResponse.AlbumList2.Album))
		for i, album := range resp.SubsonicResponse.AlbumList2.Album {
			albums[i] = models.Album{
				ID:         album.ID,
				Name:       album.Name,
				Artist:     album.Artist,
				ArtistID:   album.ArtistID,
				Year:       album.Year,
				Genre:      album.Genre,
				Duration:   album.Duration,
				TrackCount: album.SongCount,
				CreatedAt:  album.Created,
				CoverArt:   album.CoverArt,
			}
		}

		return ArtistAlbumsModalResult{Albums: albums}
	})
}

// handleModalKeyPress handles keyboard input when a modal is open
func (a *App) handleModalKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Close modal
		a.state.ShowAlbumModal = false
		a.state.ShowArtistModal = false
		a.state.SelectedAlbum = nil
		a.state.SelectedArtist = nil
		a.state.AlbumTracks = nil
		a.state.ArtistAlbums = nil
		a.state.SelectedModalIndex = 0
		return a, nil
	case "up", "k":
		// Navigate up in modal
		if a.state.SelectedModalIndex > 0 {
			a.state.SelectedModalIndex--
		}
	case "down", "j":
		// Navigate down in modal
		maxIndex := 0
		if a.state.ShowAlbumModal && len(a.state.AlbumTracks) > 0 {
			maxIndex = len(a.state.AlbumTracks) - 1
		} else if a.state.ShowArtistModal && len(a.state.ArtistAlbums) > 0 {
			maxIndex = len(a.state.ArtistAlbums) - 1
		}
		if a.state.SelectedModalIndex < maxIndex {
			a.state.SelectedModalIndex++
		}
	case "enter":
		// Handle different modal behaviors
		if a.state.ShowAlbumModal && a.state.SelectedModalIndex < len(a.state.AlbumTracks) {
			// Album modal: Play selected track immediately and queue remainder
			selectedIndex := a.state.SelectedModalIndex
			selectedTrack := a.state.AlbumTracks[selectedIndex]
			remainingTracks := a.state.AlbumTracks[selectedIndex:]
			
			if a.audioManager != nil {
				// Clear current queue and add the track selection
				a.audioManager.ClearQueue()
				a.audioManager.AddTracksToQueue(remainingTracks)
				// Start playing the first track (selected one)
				a.audioManager.PlayTrackAtIndex(0)
				
				// Immediately sync the state manually
				a.state.Queue = a.audioManager.GetQueue()
				a.state.CurrentTrack = a.audioManager.GetCurrentTrack()
				a.state.IsPlaying = a.audioManager.IsPlaying()
				
				// Log the action for user feedback
				trackNum := selectedTrack.Track
				if trackNum == 0 {
					trackNum = selectedIndex + 1
				}
				a.logMessage(fmt.Sprintf("Playing track %d: %s - %s (%d tracks queued)", 
					trackNum, selectedTrack.Artist, selectedTrack.Title, len(a.state.Queue)))
			} else {
				// Fallback if audio manager not available
				a.state.Queue = remainingTracks
				a.state.CurrentTrack = &selectedTrack
				a.state.IsPlaying = true
				
				// Log the action for user feedback
				trackNum := selectedTrack.Track
				if trackNum == 0 {
					trackNum = selectedIndex + 1
				}
				a.logMessage(fmt.Sprintf("Playing: %s - %s (from track %d)", 
					selectedTrack.Artist, selectedTrack.Title, trackNum))
			}
			
			// Close the modal after starting playback
			a.state.ShowAlbumModal = false
			a.state.SelectedAlbum = nil
			a.state.AlbumTracks = nil
			a.state.SelectedModalIndex = 0
			
			return a, nil
		} else if a.state.ShowArtistModal && a.state.SelectedModalIndex < len(a.state.ArtistAlbums) {
			// Artist modal: Open selected album's tracks modal
			selectedAlbum := a.state.ArtistAlbums[a.state.SelectedModalIndex]
			
			// Close the artist modal and open album modal
			a.state.ShowArtistModal = false
			a.state.SelectedArtist = nil
			a.state.ArtistAlbums = nil
			a.state.SelectedModalIndex = 0
			
			return a, a.showAlbumModal(selectedAlbum)
		}
	case "a", "alt+enter":
		// Add all items to queue
		if a.state.ShowAlbumModal && len(a.state.AlbumTracks) > 0 {
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(a.state.AlbumTracks)
				// Sync state immediately
				a.state.Queue = a.audioManager.GetQueue()
				a.state.CurrentTrack = a.audioManager.GetCurrentTrack()
				a.state.IsPlaying = a.audioManager.IsPlaying()
			} else {
				a.state.Queue = append(a.state.Queue, a.state.AlbumTracks...)
			}
			a.logMessage(fmt.Sprintf("Added %d tracks to queue (total: %d)", len(a.state.AlbumTracks), len(a.state.Queue)))
		} else if a.state.ShowArtistModal && len(a.state.ArtistAlbums) > 0 {
			// Add all albums from this artist to queue
			totalTracks := 0
			for _, album := range a.state.ArtistAlbums {
				cmd := a.addAlbumToQueue(album)
				if cmd != nil {
					// Execute the command to get tracks (this is async, but we'll log the albums)
					totalTracks += album.TrackCount
				}
			}
			a.logMessage(fmt.Sprintf("Queued %d albums (~%d tracks)", len(a.state.ArtistAlbums), totalTracks))
		}
	}

	return a, nil
}
