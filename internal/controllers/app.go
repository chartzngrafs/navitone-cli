package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"navitone-cli/internal/audio"
	"navitone-cli/internal/config"
	"navitone-cli/internal/models"
	"navitone-cli/internal/utils"
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

// setupDebugLogging sets up file logging for debug output
func setupDebugLogging() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // If we can't get home dir, skip logging
	}
	
	tmpDir := filepath.Join(homeDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return // If we can't create tmp dir, skip logging
	}
	
	logFile := filepath.Join(tmpDir, "navitone.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return // If we can't open log file, skip logging
	}
	
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("=== Navitone Debug Session Started ===")
}

// NewApp creates a new application instance
func NewApp() *App {
	// Set up debug logging to ~/tmp/navitone.log
	setupDebugLogging()
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	state := &models.AppState{
		CurrentTab: models.HomeTab,
		Volume:     cfg.Audio.Volume,
		Queue:      make([]models.Track, 0),
		ConfigForm: models.NewConfigFormState(cfg),
		Albums:      make([]models.Album, 0),
		Artists:     make([]models.Artist, 0),
		Playlists:   make([]models.Playlist, 0),
		LogMessages: make([]string, 0),
		
		// Initialize Home tab state
		HomeSelectedSection:  0, // Start with Recently Added section
		HomeSelectedIndex:    0,
		RecentlyAddedAlbums:  make([]models.Album, 0),
		TopArtistsByPlays:    make([]models.Artist, 0),
		MostPlayedAlbums:     make([]models.Album, 0),
		TopTracks:            make([]models.Track, 0),
		LoadingHomeData:      false,
	}


    app := &App{
        state: state,
        view:  views.NewMainView(state, cfg.UI.Theme, cfg.UI.AccentIndex),
    }

    // Initialize Navidrome client if config is valid
    app.initializeNavidromeClient()

    // Initialize scrobbling manager
    app.scrobbler = scrobbling.NewManager(cfg)
    if app.navidromeClient != nil {
        app.scrobbler.AttachNavidromeClient(app.navidromeClient)
    }

    // Detect server scrobbling capability
    app.updateServerScrobbleStatus()

	// Initialize audio manager
	if app.navidromeClient != nil {
		audioManager, err := audio.NewManager(app.navidromeClient, app.scrobbler)
		if err == nil {
			app.audioManager = audioManager
			// Set up callback to update app state when audio changes
			audioManager.SetStateCallback(app.updateAudioState)
			// Set up callback for log messages
			audioManager.SetLogCallback(app.logMessage)
			// Set initial volume from config
			audioManager.SetVolume(float64(cfg.Audio.Volume) / 100.0)
			app.logMessage("Audio manager initialized successfully")
		} else {
			app.logMessage(fmt.Sprintf("Failed to create audio manager: %v", err))
		}
	} else {
		app.logMessage("Audio manager not initialized - Navidrome client is nil (check config)")
	}

	app.logMessage("Navitone started successfully")
	
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

		// Update shuffle state
		a.state.IsShuffleMode = a.audioManager.IsShuffleEnabled()

		// Update position from audio manager
		a.state.Position = a.audioManager.GetPosition()
	}
}

// logMessage adds a message to the app's log area
func (a *App) logMessage(message string) {
	a.state.AddLogMessage(message)
}

// cleanup handles graceful shutdown of all resources
func (a *App) cleanup() tea.Cmd {
	if a.audioManager != nil {
		a.audioManager.Close()
	}
	return tea.Quit
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	// Load initial data for the current tab
	if a.state.CurrentTab == models.HomeTab && a.navidromeClient != nil {
		return a.loadHomeData()
	}
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle modal navigation first
		if a.state.ShowAlbumModal || a.state.ShowArtistModal || a.state.ShowPlaylistModal || a.state.ShowSearchModal || a.state.ShowSortModal {
			return a.handleModalKeyPress(msg)
		}
		return a.handleKeyPress(msg)
	case tea.MouseMsg:
		return a.handleMouseEvent(msg)
	case tea.WindowSizeMsg:
		// Debug: ignore invalid window size messages that might be causing the header to disappear
		if msg.Width > 0 && msg.Height > 0 {
			a.view.SetSize(msg.Width, msg.Height)
		}
		return a, nil
	case ConnectionTestResult:
		// Handle connection test result
		cf := a.state.ConfigForm
		cf.TestingConnection = false
		cf.ConnectionStatus = msg.Message
		// Reinitialize client if connection was successful
        if msg.Success {
            a.initializeNavidromeClient()
            if a.scrobbler != nil && a.navidromeClient != nil {
                a.scrobbler.AttachNavidromeClient(a.navidromeClient)
            }
            // Refresh server scrobble status after reconnection
            a.updateServerScrobbleStatus()
        }
        return a, nil
	case AlbumsLoadResult:
		// Handle albums load result
		a.state.LoadingAlbums = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			// Replace with all albums
			a.state.Albums = msg.Albums
			a.state.LoadingError = ""
		}
		return a, nil
	case AlbumsSortResult:
		// Handle albums sort result
		a.state.LoadingAlbums = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
			a.logMessage(fmt.Sprintf("Sort failed: %s", msg.Error.Error()))
		} else if msg.UseInMemorySort {
			// Fallback to in-memory sorting for unsupported API sorts (like year)
			a.sortAlbumsInMemory(msg.SortBy)
			a.logMessage(fmt.Sprintf("Sorted by %s (in-memory)", msg.SortBy))
		} else {
			// Use API-sorted results
			a.state.Albums = msg.Albums
			a.state.SelectedAlbumIndex = 0
			a.state.LoadingError = ""
			a.logMessage(fmt.Sprintf("Sorted by %s", msg.SortBy))
		}
		return a, nil
	case ArtistsSortResult:
		// Handle artists sort result  
		if msg.UseInMemorySort {
			a.sortArtistsInMemory(msg.SortBy)
			a.logMessage(fmt.Sprintf("Sorted artists by %s", msg.SortBy))
		}
		return a, nil
	case PlaylistsSortResult:
		// Handle playlists sort result
		if msg.UseInMemorySort {
			a.sortPlaylistsInMemory(msg.SortBy) 
			a.logMessage(fmt.Sprintf("Sorted playlists by %s", msg.SortBy))
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
	case PlaylistsLoadResult:
		// Handle playlists load result
		a.state.LoadingPlaylists = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.Playlists = msg.Playlists
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
				// State will be updated via the audio manager callback
				a.logMessage(fmt.Sprintf("Added album to queue (%d tracks)", len(msg.Tracks)))
			} else {
				a.state.Queue = append(a.state.Queue, msg.Tracks...)
				a.logMessage(fmt.Sprintf("Added album to queue (%d tracks, total: %d)", len(msg.Tracks), len(a.state.Queue)))
			}
			a.state.LoadingError = ""
		}
		return a, nil
	case PlaylistTracksQueueResult:
		// Handle playlist tracks load result and add to queue
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			// Add all tracks to queue
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(msg.Tracks)
				// State will be updated via the audio manager callback
				a.logMessage(fmt.Sprintf("Added playlist to queue (%d tracks)", len(msg.Tracks)))
			} else {
				a.state.Queue = append(a.state.Queue, msg.Tracks...)
				a.logMessage(fmt.Sprintf("Added playlist to queue (%d tracks, total: %d)", len(msg.Tracks), len(a.state.Queue)))
			}
			a.state.LoadingError = ""
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
				// State will be updated via the audio manager callback
				a.logMessage(fmt.Sprintf("Added artist tracks to queue (%d tracks)", len(msg.Tracks)))
			} else {
				a.state.Queue = append(a.state.Queue, msg.Tracks...)
				a.logMessage(fmt.Sprintf("Added artist tracks to queue (%d tracks, total: %d)", len(msg.Tracks), len(a.state.Queue)))
			}
			a.state.LoadingError = ""
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
	case HomeDataLoadResult:
		// Handle home data load result
		a.state.LoadingHomeData = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.RecentlyAddedAlbums = msg.RecentlyAdded
			a.state.TopArtistsByPlays = msg.TopArtists
			a.state.MostPlayedAlbums = msg.MostPlayed
			a.state.TopTracks = msg.TopTracks
			a.state.LoadingError = ""
			a.logMessage("Home tab data loaded successfully")
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
	case PlaylistTracksModalResult:
		// Handle playlist tracks load for modal display
		a.state.LoadingModalContent = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.PlaylistTracks = msg.Tracks
			a.state.SelectedModalIndex = 0
			a.state.LoadingError = ""
		}
		return a, nil
	case SearchResult:
		// Handle search result
		a.state.LoadingSearchResults = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			a.state.SearchResults = msg.Results
			a.state.SelectedSearchIndex = 0
			a.state.LoadingError = ""
		}
		return a, nil
	case SearchMoreResult:
		// Handle search more result
		a.state.LoadingSearchResults = false
		if msg.Error != nil {
			a.state.LoadingError = msg.Error.Error()
		} else {
			// Append new results to existing ones
			switch msg.Section {
			case "artists":
				a.state.SearchResults.Artists = append(a.state.SearchResults.Artists, msg.Artists...)
			case "albums":
				a.state.SearchResults.Albums = append(a.state.SearchResults.Albums, msg.Albums...)
			case "tracks":
				a.state.SearchResults.Tracks = append(a.state.SearchResults.Tracks, msg.Tracks...)
			}
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
	// Handle global player controls FIRST (before tab-specific handlers)
	switch msg.String() {
	case " ":
		// Global: Space bar Play/Pause toggle
		if a.audioManager != nil {
			err := a.audioManager.TogglePlayPause()
			if err != nil {
				a.logMessage(fmt.Sprintf("Play/Pause error: %v", err))
			}
			// Let normal Bubble Tea update cycle handle state sync to prevent race conditions
		} else {
			a.state.IsPlaying = !a.state.IsPlaying
		}
		return a, nil
	case "alt+right":
		// Global: Alt+Right arrow - Next track
		if a.audioManager != nil {
			err := a.audioManager.NextTrack()
			if err != nil {
				a.logMessage(fmt.Sprintf("Next track error: %v", err))
			}
		}
		return a, nil
	case "alt+left":
		// Global: Alt+Left arrow - Previous track
		if a.audioManager != nil {
			err := a.audioManager.PreviousTrack()
			if err != nil {
				a.logMessage(fmt.Sprintf("Previous track error: %v", err))
			}
		}
		return a, nil
	case "alt+s":
		// Global: Alt+S - Toggle shuffle
		if a.audioManager != nil {
			a.audioManager.ToggleShuffle()
			// Let normal Bubble Tea update cycle handle state sync to prevent race conditions
			a.logMessage("Shuffle toggled")
		} else {
			a.state.IsShuffleMode = !a.state.IsShuffleMode
		}
		return a, nil
	case "right":
		// Global: Right arrow - Seek forward (scrub)
		if a.audioManager != nil {
			err := a.audioManager.SeekForward(10) // 10 seconds forward
			if err != nil {
				a.logMessage(fmt.Sprintf("Seek forward error: %v", err))
			}
		}
		return a, nil
	case "left":
		// Global: Left arrow - Seek backward (scrub)
		if a.audioManager != nil {
			err := a.audioManager.SeekBackward(10) // 10 seconds backward
			if err != nil {
				a.logMessage(fmt.Sprintf("Seek backward error: %v", err))
			}
		}
		return a, nil
	case "shift+up":
		// Global: Volume up
		if a.audioManager != nil {
			currentVolume := a.audioManager.GetVolume()
			newVolume := currentVolume + 0.05 // Increase by 5%
			if newVolume > 1.0 {
				newVolume = 1.0
			}
			a.audioManager.SetVolume(newVolume)
			a.state.Volume = int(newVolume * 100) // Sync UI state
		}
		return a, nil
	case "shift+down":
		// Global: Volume down
		if a.audioManager != nil {
			currentVolume := a.audioManager.GetVolume()
			newVolume := currentVolume - 0.05 // Decrease by 5%
			if newVolume < 0.0 {
				newVolume = 0.0
			}
			a.audioManager.SetVolume(newVolume)
			a.state.Volume = int(newVolume * 100) // Sync UI state
		}
		return a, nil
	case "shift+f", "F":
		// Global: Shift+F - Open search modal
		a.state.ShowSearchModal = true
		a.state.SearchQuery = ""
		a.state.SearchResults = models.SearchResults{}
		a.state.SelectedSearchIndex = 0
		a.state.LoadingSearchResults = false
		a.state.SearchArtistsOffset = 0
		a.state.SearchAlbumsOffset = 0
		a.state.SearchTracksOffset = 0
		return a, nil
	case "shift+s", "S":
		// Global: Shift+S - Open sort modal (only in sortable contexts)
		if a.state.CurrentTab == models.AlbumsTab || 
		   a.state.CurrentTab == models.ArtistsTab || 
		   a.state.CurrentTab == models.PlaylistsTab {
			a.state.ShowSortModal = true
			a.state.SelectedSortIndex = 0
			// Set context based on current tab
			switch a.state.CurrentTab {
			case models.AlbumsTab:
				a.state.CurrentSortContext = "albums"
			case models.ArtistsTab:
				a.state.CurrentSortContext = "artists"
			case models.PlaylistsTab:
				a.state.CurrentSortContext = "playlists"
			}
		}
		return a, nil
	case "shift+c", "C":
		// Global: Shift+C - Launch Cava audio visualizer in new terminal
		if err := utils.LaunchCavaInTerminal(); err != nil {
			a.logMessage(fmt.Sprintf("Failed to launch Cava: %v", err))
		} else {
			a.logMessage("Launched Cava audio visualizer")
		}
		return a, nil
	}

	// Handle config form input if in config tab
	if a.state.CurrentTab == models.ConfigTab {
		return a.handleConfigKeyPress(msg)
	}

	// Handle content browsing tabs
	if a.state.CurrentTab == models.HomeTab {
		return a.handleHomeKeyPress(msg)
	}
	if a.state.CurrentTab == models.AlbumsTab {
		return a.handleAlbumsKeyPress(msg)
	}
	if a.state.CurrentTab == models.ArtistsTab {
		return a.handleArtistsKeyPress(msg)
	}
	if a.state.CurrentTab == models.PlaylistsTab {
		return a.handlePlaylistsKeyPress(msg)
	}
	if a.state.CurrentTab == models.QueueTab {
		return a.handleQueueKeyPress(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
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
		return a, a.cleanup()
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
	case "up":
		a.moveConfigField(-1)
	case "down":
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

// updateServerScrobbleStatus checks Navidrome for server-side scrobbling status
func (a *App) updateServerScrobbleStatus() {
    if a.navidromeClient == nil || a.state == nil || a.state.ConfigForm == nil {
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    caps, err := a.navidromeClient.GetScrobblingCapabilities(ctx)
    if err != nil || caps == nil {
        a.state.ConfigForm.ServerScrobblingDetected = false
        a.state.ConfigForm.ServerScrobblingEnabled = false
        return
    }

    a.state.ConfigForm.ServerScrobblingDetected = true
    a.state.ConfigForm.ServerScrobblingEnabled = caps.UserScrobblingEnabled
}

// handleTabChange handles actions when switching tabs
func (a *App) handleTabChange() tea.Cmd {
    // Load data when entering certain tabs
    switch a.state.CurrentTab {
	case models.HomeTab:
		if a.navidromeClient != nil && !a.state.LoadingHomeData {
			// Load home data if we don't have any or if it's been a while
			if len(a.state.RecentlyAddedAlbums) == 0 && len(a.state.TopArtistsByPlays) == 0 {
				return a.loadHomeData()
			}
		}
	case models.AlbumsTab:
		if len(a.state.Albums) == 0 && a.navidromeClient != nil && !a.state.LoadingAlbums {
			return a.loadAlbums()
		}
	case models.ArtistsTab:
		if len(a.state.Artists) == 0 && a.navidromeClient != nil && !a.state.LoadingArtists {
			return a.loadArtists()
		}
    case models.PlaylistsTab:
        if len(a.state.Playlists) == 0 && a.navidromeClient != nil && !a.state.LoadingPlaylists {
            return a.loadPlaylists()
        }
    case models.ConfigTab:
        // Refresh server scrobbling status on entering Config tab
        a.updateServerScrobbleStatus()
    }
    return nil
}

// handleHomeKeyPress handles keyboard input for the home tab
func (a *App) handleHomeKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up":
		// Navigate up through all items across all sections
		a.moveHomeSelectionUp()
	case "down":
		// Navigate down through all items across all sections
		a.moveHomeSelectionDown()
	case "pgup":
		// Jump to previous section or move up significantly within current section
		a.moveHomeSelectionPageUp()
	case "pgdown":
		// Jump to next section or move down significantly within current section
		a.moveHomeSelectionPageDown()
	case "enter":
		// Handle selection based on current section
		return a.handleHomeSelection(false) // false = play and queue
	case "shift+enter":
		// Handle selection with queue only
		return a.handleHomeSelection(true) // true = queue only
	case "r":
		// Refresh home data
		return a, a.loadHomeData()
	}

	return a, nil
}

// getTotalHomeItems returns the total number of items across all home sections
func (a *App) getTotalHomeItems() int {
	return len(a.state.RecentlyAddedAlbums) + len(a.state.TopArtistsByPlays) + 
		   len(a.state.MostPlayedAlbums) + len(a.state.TopTracks)
}

// getGlobalHomeIndex returns the global index across all sections
func (a *App) getGlobalHomeIndex() int {
	globalIndex := 0
	
	// Add items from sections before the current one
	if a.state.HomeSelectedSection > 0 {
		globalIndex += len(a.state.RecentlyAddedAlbums)
	}
	if a.state.HomeSelectedSection > 1 {
		globalIndex += len(a.state.TopArtistsByPlays)
	}
	if a.state.HomeSelectedSection > 2 {
		globalIndex += len(a.state.MostPlayedAlbums)
	}
	
	// Add the current section index
	globalIndex += a.state.HomeSelectedIndex
	
	return globalIndex
}

// setHomeSelectionFromGlobalIndex sets the section and index from a global index
func (a *App) setHomeSelectionFromGlobalIndex(globalIndex int) {
	currentIndex := 0
	
	// Section 0: Recently Added Albums
	if globalIndex < currentIndex + len(a.state.RecentlyAddedAlbums) {
		a.state.HomeSelectedSection = 0
		a.state.HomeSelectedIndex = globalIndex - currentIndex
		return
	}
	currentIndex += len(a.state.RecentlyAddedAlbums)
	
	// Section 1: Top Artists
	if globalIndex < currentIndex + len(a.state.TopArtistsByPlays) {
		a.state.HomeSelectedSection = 1
		a.state.HomeSelectedIndex = globalIndex - currentIndex
		return
	}
	currentIndex += len(a.state.TopArtistsByPlays)
	
	// Section 2: Most Played Albums
	if globalIndex < currentIndex + len(a.state.MostPlayedAlbums) {
		a.state.HomeSelectedSection = 2
		a.state.HomeSelectedIndex = globalIndex - currentIndex
		return
	}
	currentIndex += len(a.state.MostPlayedAlbums)
	
	// Section 3: Top Tracks
	if globalIndex < currentIndex + len(a.state.TopTracks) {
		a.state.HomeSelectedSection = 3
		a.state.HomeSelectedIndex = globalIndex - currentIndex
		return
	}
}

// moveHomeSelectionUp moves selection up across all sections
func (a *App) moveHomeSelectionUp() {
	globalIndex := a.getGlobalHomeIndex()
	if globalIndex > 0 {
		a.setHomeSelectionFromGlobalIndex(globalIndex - 1)
	}
}

// moveHomeSelectionDown moves selection down across all sections
func (a *App) moveHomeSelectionDown() {
	globalIndex := a.getGlobalHomeIndex()
	totalItems := a.getTotalHomeItems()
	if globalIndex < totalItems - 1 {
		a.setHomeSelectionFromGlobalIndex(globalIndex + 1)
	}
}

// moveHomeSelectionPageUp moves selection up by sections or large jumps
func (a *App) moveHomeSelectionPageUp() {
	// Jump to the beginning of the current section, or previous section if already at beginning
	if a.state.HomeSelectedIndex == 0 && a.state.HomeSelectedSection > 0 {
		// Jump to previous section
		a.state.HomeSelectedSection--
		switch a.state.HomeSelectedSection {
		case 0:
			a.state.HomeSelectedIndex = len(a.state.RecentlyAddedAlbums) - 1
		case 1:
			a.state.HomeSelectedIndex = len(a.state.TopArtistsByPlays) - 1
		case 2:
			a.state.HomeSelectedIndex = len(a.state.MostPlayedAlbums) - 1
		case 3:
			a.state.HomeSelectedIndex = len(a.state.TopTracks) - 1
		}
		if a.state.HomeSelectedIndex < 0 {
			a.state.HomeSelectedIndex = 0
		}
	} else {
		// Jump to beginning of current section
		a.state.HomeSelectedIndex = 0
	}
}

// moveHomeSelectionPageDown moves selection down by sections or large jumps  
func (a *App) moveHomeSelectionPageDown() {
	// Jump to the end of the current section, or next section if already at end
	currentSectionSize := 0
	switch a.state.HomeSelectedSection {
	case 0:
		currentSectionSize = len(a.state.RecentlyAddedAlbums)
	case 1:
		currentSectionSize = len(a.state.TopArtistsByPlays)
	case 2:
		currentSectionSize = len(a.state.MostPlayedAlbums)
	case 3:
		currentSectionSize = len(a.state.TopTracks)
	}
	
	if a.state.HomeSelectedIndex == currentSectionSize-1 && a.state.HomeSelectedSection < 3 {
		// Jump to next section
		a.state.HomeSelectedSection++
		a.state.HomeSelectedIndex = 0
	} else {
		// Jump to end of current section
		a.state.HomeSelectedIndex = currentSectionSize - 1
		if a.state.HomeSelectedIndex < 0 {
			a.state.HomeSelectedIndex = 0
		}
	}
}

// handleHomeSelection handles when an item is selected in the home tab
func (a *App) handleHomeSelection(queueOnly bool) (tea.Model, tea.Cmd) {
	switch a.state.HomeSelectedSection {
	case 0: // Recently Added Albums
		if a.state.HomeSelectedIndex < len(a.state.RecentlyAddedAlbums) {
			album := a.state.RecentlyAddedAlbums[a.state.HomeSelectedIndex]
			return a, a.showAlbumModal(album)
		}
	case 1: // Top Artists
		if a.state.HomeSelectedIndex < len(a.state.TopArtistsByPlays) {
			artist := a.state.TopArtistsByPlays[a.state.HomeSelectedIndex]
			return a, a.showArtistModal(artist)
		}
	case 2: // Most Played Albums
		if a.state.HomeSelectedIndex < len(a.state.MostPlayedAlbums) {
			album := a.state.MostPlayedAlbums[a.state.HomeSelectedIndex]
			return a, a.showAlbumModal(album)
		}
	case 3: // Top Tracks
		if a.state.HomeSelectedIndex < len(a.state.TopTracks) {
			track := a.state.TopTracks[a.state.HomeSelectedIndex]
			if queueOnly {
				return a, a.addTrackToQueue(track)
			} else {
				// Play track and queue remaining
				remainingTracks := a.state.TopTracks[a.state.HomeSelectedIndex:]
				if a.audioManager != nil {
					a.audioManager.ClearQueue()
					a.audioManager.AddTracksToQueue(remainingTracks)
					a.audioManager.PlayTrackAtIndex(0)
					a.logMessage(fmt.Sprintf("Playing: %s - %s (%d tracks queued)", 
						track.Artist, track.Title, len(remainingTracks)))
				} else {
					a.state.Queue = remainingTracks
					a.state.CurrentTrack = &track
					a.state.IsPlaying = true
					a.logMessage(fmt.Sprintf("Playing: %s - %s", track.Artist, track.Title))
				}
			}
		}
	}
	return a, nil
}

// handleAlbumsKeyPress handles keyboard input for the albums tab
func (a *App) handleAlbumsKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	
	switch msg.String() {
	case "ctrl+c", "q":
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up":
		if a.state.SelectedAlbumIndex > 0 {
			a.state.SelectedAlbumIndex--
		}
	case "down":
		if a.state.SelectedAlbumIndex < len(a.state.Albums)-1 {
			a.state.SelectedAlbumIndex++
		}
	case "pgup":
		// Move up by 25 items
		a.state.SelectedAlbumIndex -= 25
		if a.state.SelectedAlbumIndex < 0 {
			a.state.SelectedAlbumIndex = 0
		}
	case "pgdown":
		// Move down by 25 items
		a.state.SelectedAlbumIndex += 25
		if a.state.SelectedAlbumIndex >= len(a.state.Albums) {
			a.state.SelectedAlbumIndex = len(a.state.Albums) - 1
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

// loadAlbums loads all albums from Navidrome library
func (a *App) loadAlbums() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingAlbums = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for all albums
		defer cancel()

		// Load all albums by setting a very high limit
		resp, err := a.navidromeClient.GetAlbums(ctx, 10000, 0)
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
				PlayCount:  album.PlayCount,
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

		// Convert Navidrome artists to our model and calculate play counts
		artistPlayCounts := make(map[string]int)
		var artists []models.Artist
		
		// First, collect all artists
		for _, index := range resp.SubsonicResponse.Artists.Index {
			for _, artist := range index.Artist {
				artists = append(artists, models.Artist{
					ID:         artist.ID,
					Name:       artist.Name,
					AlbumCount: artist.AlbumCount,
					PlayCount:  0, // Will be calculated below
					StarredAt:  artist.Starred,
				})
				artistPlayCounts[artist.ID] = 0
			}
		}
		
		// Get all albums to aggregate play counts per artist
		// Use alphabeticalByName to get ALL albums, not just frequent ones
		allAlbumsResp, err := a.navidromeClient.GetAlbumsByType(ctx, "alphabeticalByName", 1000, 0)
		if err == nil {
			// Aggregate play counts for each artist from their albums
			for _, album := range allAlbumsResp.SubsonicResponse.AlbumList2.Album {
				if count, exists := artistPlayCounts[album.ArtistID]; exists {
					artistPlayCounts[album.ArtistID] = count + album.PlayCount
				}
			}
		}
		
		// Update artists with aggregated play counts
		for i := range artists {
			if count, exists := artistPlayCounts[artists[i].ID]; exists {
				artists[i].PlayCount = count
			}
		}

		return ArtistsLoadResult{Artists: artists}
	})
}

// loadPlaylists loads playlists from Navidrome
func (a *App) loadPlaylists() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingPlaylists = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := a.navidromeClient.GetPlaylists(ctx)
		if err != nil {
			return PlaylistsLoadResult{Error: err}
		}

		// Convert Navidrome playlists to our model
		playlists := make([]models.Playlist, len(resp.SubsonicResponse.Playlists.Playlist))
		for i, playlist := range resp.SubsonicResponse.Playlists.Playlist {
			playlists[i] = models.Playlist{
				ID:        playlist.ID,
				Name:      playlist.Name,
				Comment:   playlist.Comment,
				Owner:     playlist.Owner,
				Public:    playlist.Public,
				SongCount: playlist.SongCount,
				Duration:  playlist.Duration,
				CreatedAt: playlist.Created,
				ChangedAt: playlist.Changed,
			}
		}

		return PlaylistsLoadResult{Playlists: playlists}
	})
}

// loadHomeData loads all data needed for the home tab
func (a *App) loadHomeData() tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingHomeData = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		var homeData HomeDataLoadResult
		
		// Load Recently Added Albums
		recentResp, err := a.navidromeClient.GetAlbumsByType(ctx, "newest", 8, 0)
		if err != nil {
			homeData.Error = err
			return homeData
		}
		
		
		// Convert recently added albums
		homeData.RecentlyAdded = make([]models.Album, len(recentResp.SubsonicResponse.AlbumList2.Album))
		for i, album := range recentResp.SubsonicResponse.AlbumList2.Album {
			homeData.RecentlyAdded[i] = models.Album{
				ID:         album.ID,
				Name:       album.Name,
				Artist:     album.Artist,
				ArtistID:   album.ArtistID,
				Year:       album.Year,
				Genre:      album.Genre,
				Duration:   album.Duration,
				TrackCount: album.SongCount,
				PlayCount:  album.PlayCount,
				CreatedAt:  album.Created,
				CoverArt:   album.CoverArt,
			}
		}

		// Load Most Played Albums
		frequentResp, err := a.navidromeClient.GetAlbumsByType(ctx, "frequent", 8, 0)
		if err != nil {
			// If frequent doesn't work, try recent or newest as fallback
			frequentResp, err = a.navidromeClient.GetAlbumsByType(ctx, "recent", 8, 0)
			if err != nil {
				// Final fallback to newest
				frequentResp = recentResp // Reuse recently added as fallback
			}
		}
		
		// Convert most played albums
		homeData.MostPlayed = make([]models.Album, len(frequentResp.SubsonicResponse.AlbumList2.Album))
		for i, album := range frequentResp.SubsonicResponse.AlbumList2.Album {
			homeData.MostPlayed[i] = models.Album{
				ID:         album.ID,
				Name:       album.Name,
				Artist:     album.Artist,
				ArtistID:   album.ArtistID,
				Year:       album.Year,
				Genre:      album.Genre,
				Duration:   album.Duration,
				TrackCount: album.SongCount,
				PlayCount:  album.PlayCount,
				CreatedAt:  album.Created,
				CoverArt:   album.CoverArt,
			}
		}

		// Load Top Tracks - use tracks from most played albums since GetTopTracks returns mostly 0s
		var allTopTracks []models.Track
		
		// Get tracks from most played albums (more reliable than GetTopTracks)
		if len(homeData.MostPlayed) > 0 {
			// Get tracks from top 3 most played albums (reduced from 5 for performance)
			maxAlbums := 3
			if len(homeData.MostPlayed) < maxAlbums {
				maxAlbums = len(homeData.MostPlayed)
			}
			
			for i := 0; i < maxAlbums; i++ {
				albumTracksResp, albumErr := a.navidromeClient.GetAlbumTracks(ctx, homeData.MostPlayed[i].ID)
				if albumErr == nil {
					// Convert album tracks
					for _, song := range albumTracksResp.SubsonicResponse.SongsByGenre.Song {
						allTopTracks = append(allTopTracks, models.Track{
							ID:        song.ID,
							Title:     song.Title,
							Artist:    song.Artist,
							ArtistID:  song.ArtistID,
							Album:     song.Album,
							AlbumID:   song.AlbumID,
							Genre:     song.Genre,
							Year:      song.Year,
							Duration:  song.Duration,
							Track:     song.Track,
							Disc:      song.DiscNumber,
							Size:      song.Size,
							Suffix:    song.Suffix,
							BitRate:   song.BitRate,
							PlayCount: song.PlayCount,
							Path:      song.Path,
						})
					}
				}
			}
		}
		
		// Sort by play count (descending) and take top 10
		if len(allTopTracks) > 0 {
			for i := 0; i < len(allTopTracks)-1; i++ {
				for j := 0; j < len(allTopTracks)-i-1; j++ {
					if allTopTracks[j].PlayCount < allTopTracks[j+1].PlayCount {
						allTopTracks[j], allTopTracks[j+1] = allTopTracks[j+1], allTopTracks[j]
					}
				}
			}
			maxTracks := 10
			if len(allTopTracks) < maxTracks {
				maxTracks = len(allTopTracks)
			}
			homeData.TopTracks = allTopTracks[:maxTracks]
		} else {
			// Final fallback to random songs
			tracksResp, err := a.navidromeClient.GetSongs(ctx, 10, 0)
			if err != nil {
				homeData.Error = err
				return homeData
			}
			// Convert random tracks
			homeData.TopTracks = make([]models.Track, len(tracksResp.SubsonicResponse.SongsByGenre.Song))
			for i, song := range tracksResp.SubsonicResponse.SongsByGenre.Song {
				homeData.TopTracks[i] = models.Track{
					ID:        song.ID,
					Title:     song.Title,
					Artist:    song.Artist,
					ArtistID:  song.ArtistID,
					Album:     song.Album,
					AlbumID:   song.AlbumID,
					Genre:     song.Genre,
					Year:      song.Year,
					Duration:  song.Duration,
					Track:     song.Track,
					Disc:      song.DiscNumber,
					Size:      song.Size,
					Suffix:    song.Suffix,
					BitRate:   song.BitRate,
					PlayCount: song.PlayCount,
					Path:      song.Path,
				}
			}
		}

		// Load Top Artists (aggregate play counts from albums)
		artistsResp, err := a.navidromeClient.GetArtists(ctx)
		if err != nil {
			homeData.Error = err
			return homeData
		}
		
		// Create a map to aggregate play counts per artist
		artistPlayCounts := make(map[string]int)
		var allArtists []models.Artist
		
		// First, collect all artists
		for _, index := range artistsResp.SubsonicResponse.Artists.Index {
			for _, artist := range index.Artist {
				allArtists = append(allArtists, models.Artist{
					ID:         artist.ID,
					Name:       artist.Name,
					AlbumCount: artist.AlbumCount,
					PlayCount:  0, // Will be calculated below
					StarredAt:  artist.Starred,
				})
				artistPlayCounts[artist.ID] = 0
			}
		}
		
		// Get all albums to aggregate play counts per artist
		// We'll use the "frequent" albums to get play count data efficiently
		allAlbumsResp, err := a.navidromeClient.GetAlbumsByType(ctx, "frequent", 200, 0)
		if err == nil {
			// Aggregate play counts for each artist from their albums
			for _, album := range allAlbumsResp.SubsonicResponse.AlbumList2.Album {
				if count, exists := artistPlayCounts[album.ArtistID]; exists {
					artistPlayCounts[album.ArtistID] = count + album.PlayCount
				}
			}
		}
		
		// Update artists with aggregated play counts
		for i := range allArtists {
			if count, exists := artistPlayCounts[allArtists[i].ID]; exists {
				allArtists[i].PlayCount = count
			}
		}
		
		// Sort artists by play count (descending), fallback to album count
		for i := 0; i < len(allArtists)-1; i++ {
			for j := 0; j < len(allArtists)-i-1; j++ {
				// Primary sort by play count, secondary by album count
				leftScore := allArtists[j].PlayCount*1000 + allArtists[j].AlbumCount
				rightScore := allArtists[j+1].PlayCount*1000 + allArtists[j+1].AlbumCount
				if leftScore < rightScore {
					allArtists[j], allArtists[j+1] = allArtists[j+1], allArtists[j]
				}
			}
		}
		
		// Take top 5 artists
		maxArtists := 5
		if len(allArtists) < maxArtists {
			maxArtists = len(allArtists)
		}
		homeData.TopArtists = allArtists[:maxArtists]

		return homeData
	})
}

// HomeDataLoadResult represents the result of loading home tab data
type HomeDataLoadResult struct {
	RecentlyAdded []models.Album
	MostPlayed    []models.Album
	TopTracks     []models.Track
	TopArtists    []models.Artist
	Error         error
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

// addPlaylistToQueue adds all tracks from a playlist to the queue
func (a *App) addPlaylistToQueue(playlist models.Playlist) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if a.navidromeClient == nil {
				return PlaylistTracksQueueResult{Error: fmt.Errorf("navidrome client not initialized")}
			}

			// Add timeout context to prevent hanging
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Fetch actual tracks from the playlist
			resp, err := a.navidromeClient.GetPlaylistTracks(ctx, playlist.ID)
			if err != nil {
				return PlaylistTracksQueueResult{Error: fmt.Errorf("failed to queue playlist tracks: %w", err)}
			}

			// Check if response structure is valid
			if resp == nil {
				return PlaylistTracksQueueResult{Error: fmt.Errorf("received null response")}
			}
			
			entryCount := len(resp.SubsonicResponse.Playlist.Entry)
			if entryCount == 0 {
				return PlaylistTracksQueueResult{Tracks: []models.Track{}}
			}

			// Add a safety limit to prevent massive allocations
			if entryCount > 10000 {
				return PlaylistTracksQueueResult{Error: fmt.Errorf("playlist too large: %d tracks", entryCount)}
			}

			// Convert Navidrome songs to our model
			tracks := make([]models.Track, entryCount)
			for i, song := range resp.SubsonicResponse.Playlist.Entry {
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

			return PlaylistTracksQueueResult{Tracks: tracks}
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
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up":
		if a.state.SelectedArtistIndex > 0 {
			a.state.SelectedArtistIndex--
		}
	case "down":
		if a.state.SelectedArtistIndex < len(a.state.Artists)-1 {
			a.state.SelectedArtistIndex++
		}
	case "pgup":
		// Move up by 25 items
		a.state.SelectedArtistIndex -= 25
		if a.state.SelectedArtistIndex < 0 {
			a.state.SelectedArtistIndex = 0
		}
	case "pgdown":
		// Move down by 25 items
		a.state.SelectedArtistIndex += 25
		if a.state.SelectedArtistIndex >= len(a.state.Artists) {
			a.state.SelectedArtistIndex = len(a.state.Artists) - 1
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

// handlePlaylistsKeyPress handles keyboard input for the playlists tab
func (a *App) handlePlaylistsKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up":
		if a.state.SelectedPlaylistIndex > 0 {
			a.state.SelectedPlaylistIndex--
		}
	case "down":
		if a.state.SelectedPlaylistIndex < len(a.state.Playlists)-1 {
			a.state.SelectedPlaylistIndex++
		}
	case "pgup":
		// Move up by 25 items
		a.state.SelectedPlaylistIndex -= 25
		if a.state.SelectedPlaylistIndex < 0 {
			a.state.SelectedPlaylistIndex = 0
		}
	case "pgdown":
		// Move down by 25 items
		a.state.SelectedPlaylistIndex += 25
		if a.state.SelectedPlaylistIndex >= len(a.state.Playlists) {
			a.state.SelectedPlaylistIndex = len(a.state.Playlists) - 1
		}
	case "enter":
		// Show playlist tracks modal
		if a.state.SelectedPlaylistIndex < len(a.state.Playlists) {
			return a, a.showPlaylistModal(a.state.Playlists[a.state.SelectedPlaylistIndex])
		}
	case "alt+enter":
		// Queue entire playlist immediately (Alt+Enter)
		if a.state.SelectedPlaylistIndex < len(a.state.Playlists) {
			return a, a.addPlaylistToQueue(a.state.Playlists[a.state.SelectedPlaylistIndex])
		}
	case "a":
		// Alternative: 'A' key to queue entire playlist immediately
		if a.state.SelectedPlaylistIndex < len(a.state.Playlists) {
			return a, a.addPlaylistToQueue(a.state.Playlists[a.state.SelectedPlaylistIndex])
		}
	case "r":
		// Refresh playlists
		return a, a.loadPlaylists()
	}

	return a, nil
}

// ArtistTracksLoadResult represents the result of loading artist tracks
type ArtistTracksLoadResult struct {
	Tracks []models.Track
	Error  error
}


// addTrackToQueue adds a single track to the queue
func (a *App) addTrackToQueue(track models.Track) tea.Cmd {
	if a.audioManager != nil {
		a.audioManager.AddToQueue(track)
		// State will be updated via the audio manager callback
	} else {
		a.state.Queue = append(a.state.Queue, track)
	}
	return nil
}

// handleQueueKeyPress handles keyboard input for the queue tab
func (a *App) handleQueueKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, a.cleanup()
	case "tab":
		a.nextTab()
		return a, a.handleTabChange()
	case "shift+tab":
		a.prevTab()
		return a, a.handleTabChange()
	case "up":
		if a.state.SelectedQueueIndex > 0 {
			a.state.SelectedQueueIndex--
		}
	case "down":
		if a.state.SelectedQueueIndex < len(a.state.Queue)-1 {
			a.state.SelectedQueueIndex++
		}
	case "pgup":
		// Move up by 25 items
		a.state.SelectedQueueIndex -= 25
		if a.state.SelectedQueueIndex < 0 {
			a.state.SelectedQueueIndex = 0
		}
	case "pgdown":
		// Move down by 25 items
		a.state.SelectedQueueIndex += 25
		if a.state.SelectedQueueIndex >= len(a.state.Queue) {
			a.state.SelectedQueueIndex = len(a.state.Queue) - 1
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
	case "enter":
		// Play selected track (Enter only, Space is handled globally for play/pause)
		if a.audioManager != nil {
			if a.state.SelectedQueueIndex < len(a.state.Queue) {
				a.audioManager.PlayTrackAtIndex(a.state.SelectedQueueIndex)
			}
		} else {
			// Fallback for when audio manager is not available
			if a.state.SelectedQueueIndex < len(a.state.Queue) {
				a.state.CurrentTrack = &a.state.Queue[a.state.SelectedQueueIndex]
				a.state.IsPlaying = true
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

type PlaylistsLoadResult struct {
	Playlists []models.Playlist
	Error     error
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

type PlaylistTracksModalResult struct {
	Tracks []models.Track
	Error  error
}

type PlaylistTracksQueueResult struct {
	Tracks []models.Track
	Error  error
}

type SearchResult struct {
	Results models.SearchResults
	Error   error
}


// handleMouseEvent processes mouse input
func (a *App) handleMouseEvent(_ tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse support placeholder for Phase 2
	return a, nil
}

// nextTab switches to the next tab
func (a *App) nextTab() {
	a.state.CurrentTab = models.Tab((int(a.state.CurrentTab) + 1) % 6)
}

// prevTab switches to the previous tab
func (a *App) prevTab() {
	current := int(a.state.CurrentTab)
	if current == 0 {
		current = 6
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

// showPlaylistModal displays the playlist tracks modal
func (a *App) showPlaylistModal(playlist models.Playlist) tea.Cmd {
	a.state.ShowPlaylistModal = true
	a.state.SelectedPlaylist = &playlist
	a.state.LoadingModalContent = true
	a.state.PlaylistTracks = nil
	a.state.SelectedModalIndex = 0

	return tea.Cmd(func() tea.Msg {
		if a.navidromeClient == nil {
			return PlaylistTracksModalResult{Error: fmt.Errorf("navidrome client not initialized")}
		}

		// Add timeout context to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := a.navidromeClient.GetPlaylistTracks(ctx, playlist.ID)
		if err != nil {
			return PlaylistTracksModalResult{Error: fmt.Errorf("failed to load playlist tracks: %w", err)}
		}

		// Check if response structure is valid
		if resp == nil {
			return PlaylistTracksModalResult{Error: fmt.Errorf("received null response")}
		}
		
		entryCount := len(resp.SubsonicResponse.Playlist.Entry)
		if entryCount == 0 {
			return PlaylistTracksModalResult{Tracks: []models.Track{}}
		}

		// Add a safety limit to prevent massive allocations
		if entryCount > 10000 {
			return PlaylistTracksModalResult{Error: fmt.Errorf("playlist too large: %d tracks", entryCount)}
		}

		tracks := make([]models.Track, entryCount)
		for i, song := range resp.SubsonicResponse.Playlist.Entry {
			// Safety check to prevent infinite loops
			if i >= entryCount {
				break
			}
			
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

		return PlaylistTracksModalResult{Tracks: tracks}
	})
}

// handleModalKeyPress handles keyboard input when a modal is open
func (a *App) handleModalKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search modal first
	if a.state.ShowSearchModal {
		return a.handleSearchModalKeyPress(msg)
	}
	
	// Handle sort modal
	if a.state.ShowSortModal {
		return a.handleSortModalKeyPress(msg)
	}
	
	switch msg.String() {
	case "esc", "q":
		// Close modal
		a.state.ShowAlbumModal = false
		a.state.ShowArtistModal = false
		a.state.ShowPlaylistModal = false
		a.state.ShowSortModal = false
		a.state.SelectedAlbum = nil
		a.state.SelectedArtist = nil
		a.state.SelectedPlaylist = nil
		a.state.AlbumTracks = nil
		a.state.ArtistAlbums = nil
		a.state.PlaylistTracks = nil
		a.state.SelectedModalIndex = 0
		a.state.SelectedSortIndex = 0
		a.state.CurrentSortContext = ""
		return a, nil
	case "up":
		// Navigate up in modal
		if a.state.SelectedModalIndex > 0 {
			a.state.SelectedModalIndex--
		}
	case "down":
		// Navigate down in modal
		maxIndex := 0
		if a.state.ShowAlbumModal && len(a.state.AlbumTracks) > 0 {
			maxIndex = len(a.state.AlbumTracks) - 1
		} else if a.state.ShowArtistModal && len(a.state.ArtistAlbums) > 0 {
			maxIndex = len(a.state.ArtistAlbums) - 1
		} else if a.state.ShowPlaylistModal && len(a.state.PlaylistTracks) > 0 {
			maxIndex = len(a.state.PlaylistTracks) - 1
		}
		if a.state.SelectedModalIndex < maxIndex {
			a.state.SelectedModalIndex++
		}
	case "pgup":
		// Jump up by 10 items in modal
		a.state.SelectedModalIndex -= 10
		if a.state.SelectedModalIndex < 0 {
			a.state.SelectedModalIndex = 0
		}
	case "pgdown":
		// Jump down by 10 items in modal
		maxIndex := 0
		if a.state.ShowAlbumModal && len(a.state.AlbumTracks) > 0 {
			maxIndex = len(a.state.AlbumTracks) - 1
		} else if a.state.ShowArtistModal && len(a.state.ArtistAlbums) > 0 {
			maxIndex = len(a.state.ArtistAlbums) - 1
		} else if a.state.ShowPlaylistModal && len(a.state.PlaylistTracks) > 0 {
			maxIndex = len(a.state.PlaylistTracks) - 1
		}
		a.state.SelectedModalIndex += 10
		if a.state.SelectedModalIndex > maxIndex {
			a.state.SelectedModalIndex = maxIndex
		}
	case "enter":
		// Handle different modal behaviors
		if a.state.ShowAlbumModal && a.state.SelectedModalIndex < len(a.state.AlbumTracks) {
			// Album modal: Play selected track immediately and queue remainder
			selectedIndex := a.state.SelectedModalIndex
			selectedTrack := a.state.AlbumTracks[selectedIndex]
			remainingTracks := a.state.AlbumTracks[selectedIndex:]
			
			// Close the modal first to prevent UI interference
			a.state.ShowAlbumModal = false
			a.state.SelectedAlbum = nil
			a.state.AlbumTracks = nil
			a.state.SelectedModalIndex = 0
			
			if a.audioManager != nil {
				// Clear current queue and add the track selection
				a.audioManager.ClearQueue()
				a.audioManager.AddTracksToQueue(remainingTracks)
				// Start playing the first track (selected one)
				a.audioManager.PlayTrackAtIndex(0)
				
				// Log the action for user feedback
				trackNum := selectedTrack.Track
				if trackNum == 0 {
					trackNum = selectedIndex + 1
				}
				a.logMessage(fmt.Sprintf("Playing track %d: %s - %s (%d tracks queued)", 
					trackNum, selectedTrack.Artist, selectedTrack.Title, len(remainingTracks)))
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
		} else if a.state.ShowPlaylistModal && a.state.SelectedModalIndex < len(a.state.PlaylistTracks) {
			// Playlist modal: Play selected track immediately and queue remainder
			selectedIndex := a.state.SelectedModalIndex
			selectedTrack := a.state.PlaylistTracks[selectedIndex]
			remainingTracks := a.state.PlaylistTracks[selectedIndex:]
			
			// Close the modal first to prevent UI interference
			a.state.ShowPlaylistModal = false
			a.state.SelectedPlaylist = nil
			a.state.PlaylistTracks = nil
			a.state.SelectedModalIndex = 0
			
			if a.audioManager != nil {
				// Clear current queue and add the track selection
				a.audioManager.ClearQueue()
				a.audioManager.AddTracksToQueue(remainingTracks)
				// Start playing the first track (selected one)
				a.audioManager.PlayTrackAtIndex(0)
				
				// Log the action for user feedback
				trackNum := selectedIndex + 1
				a.logMessage(fmt.Sprintf("Playing track %d: %s - %s (%d tracks queued)", 
					trackNum, selectedTrack.Artist, selectedTrack.Title, len(remainingTracks)))
			} else {
				// Fallback if audio manager not available
				a.state.Queue = remainingTracks
				a.state.CurrentTrack = &selectedTrack
				a.state.IsPlaying = true
				
				// Log the action for user feedback
				trackNum := selectedIndex + 1
				a.logMessage(fmt.Sprintf("Playing: %s - %s (from track %d)", 
					selectedTrack.Artist, selectedTrack.Title, trackNum))
			}
			
			return a, nil
		}
	case "a", "alt+enter":
		// Add all items to queue
		if a.state.ShowAlbumModal && len(a.state.AlbumTracks) > 0 {
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(a.state.AlbumTracks)
				// State will be updated via the audio manager callback
				a.logMessage(fmt.Sprintf("Added %d tracks to queue", len(a.state.AlbumTracks)))
			} else {
				a.state.Queue = append(a.state.Queue, a.state.AlbumTracks...)
				a.logMessage(fmt.Sprintf("Added %d tracks to queue (total: %d)", len(a.state.AlbumTracks), len(a.state.Queue)))
			}
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
		} else if a.state.ShowPlaylistModal && len(a.state.PlaylistTracks) > 0 {
			// Add all playlist tracks to queue
			if a.audioManager != nil {
				a.audioManager.AddTracksToQueue(a.state.PlaylistTracks)
				// State will be updated via the audio manager callback
				a.logMessage(fmt.Sprintf("Added %d tracks to queue", len(a.state.PlaylistTracks)))
			} else {
				a.state.Queue = append(a.state.Queue, a.state.PlaylistTracks...)
				a.logMessage(fmt.Sprintf("Added %d tracks to queue (total: %d)", len(a.state.PlaylistTracks), len(a.state.Queue)))
			}
		}
	}

	return a, nil
}

// handleSearchModalKeyPress handles keyboard input in the search modal
func (a *App) handleSearchModalKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Close search modal
		a.state.ShowSearchModal = false
		a.state.SearchQuery = ""
		a.state.SearchResults = models.SearchResults{}
		a.state.SelectedSearchIndex = 0
		a.state.LoadingSearchResults = false
		a.state.SearchArtistsOffset = 0
		a.state.SearchAlbumsOffset = 0
		a.state.SearchTracksOffset = 0
		return a, nil
	case "enter":
		// Handle search result selection - Play and queue remaining
		return a.handleSearchSelection(false)
	case "shift+enter":
		// Handle search result selection - Queue only
		return a.handleSearchSelection(true)
	case "up":
		// Navigate up in search results
		if a.state.SelectedSearchIndex > 0 {
			a.state.SelectedSearchIndex--
		}
		return a, nil
	case "down":
		// Navigate down in search results
		totalResults := len(a.state.SearchResults.Artists) + len(a.state.SearchResults.Albums) + len(a.state.SearchResults.Tracks)
		
		// Add MORE buttons to total count
		if len(a.state.SearchResults.Artists) == 5 {
			totalResults++ // Add MORE artists button
		}
		if len(a.state.SearchResults.Albums) == 5 {
			totalResults++ // Add MORE albums button  
		}
		if len(a.state.SearchResults.Tracks) == 5 {
			totalResults++ // Add MORE tracks button
		}
		
		if a.state.SelectedSearchIndex < totalResults-1 {
			a.state.SelectedSearchIndex++
		}
		return a, nil
	case "backspace":
		// Remove character from search query
		if len(a.state.SearchQuery) > 0 {
			a.state.SearchQuery = a.state.SearchQuery[:len(a.state.SearchQuery)-1]
			return a, a.performSearch()
		}
		return a, nil
	default:
		// Add character to search query
		if len(msg.String()) == 1 {
			a.state.SearchQuery += msg.String()
			return a, a.performSearch()
		}
	}
	return a, nil
}

// handleSortModalKeyPress handles keyboard input in the sort modal
func (a *App) handleSortModalKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Close sort modal
		a.state.ShowSortModal = false
		a.state.SelectedSortIndex = 0
		a.state.CurrentSortContext = ""
		return a, nil
	case "enter":
		// Apply selected sort
		return a.applySorting()
	case "up":
		// Navigate up in sort options
		if a.state.SelectedSortIndex > 0 {
			a.state.SelectedSortIndex--
		}
		return a, nil
	case "down":
		// Navigate down in sort options
		availableOptions := a.getAvailableSortOptions()
		if a.state.SelectedSortIndex < len(availableOptions)-1 {
			a.state.SelectedSortIndex++
		}
		return a, nil
	}
	return a, nil
}

// getAvailableSortOptions returns sort options available for the current context
func (a *App) getAvailableSortOptions() []models.SortOption {
	var available []models.SortOption
	for _, option := range models.SortOptions {
		for _, applicable := range option.Applicable {
			if applicable == a.state.CurrentSortContext {
				available = append(available, option)
				break
			}
		}
	}
	return available
}

// applySorting applies the selected sort to the current context
func (a *App) applySorting() (tea.Model, tea.Cmd) {
	availableOptions := a.getAvailableSortOptions()
	if a.state.SelectedSortIndex >= len(availableOptions) {
		// Invalid selection, close modal
		a.state.ShowSortModal = false
		return a, nil
	}
	
	selectedOption := availableOptions[a.state.SelectedSortIndex]
	
	// Close modal first and save context for use in switch
	currentContext := a.state.CurrentSortContext
	a.state.ShowSortModal = false
	a.state.SelectedSortIndex = 0
	a.state.CurrentSortContext = ""
	a.logMessage(fmt.Sprintf("Sorting by: %s...", selectedOption.DisplayName))
	
	// Apply sorting based on context and option - return command for async operation
	switch currentContext {
	case "albums":
		return a, a.sortAlbumsAsync(selectedOption.ID)
	case "artists":
		return a, a.sortArtistsAsync(selectedOption.ID)
	case "playlists":
		return a, a.sortPlaylistsAsync(selectedOption.ID)
	}
	
	return a, nil
}

// performSearch performs the actual search with a timeout
func (a *App) performSearch() tea.Cmd {
	if a.navidromeClient == nil || len(a.state.SearchQuery) == 0 {
		// Clear results if no query
		a.state.SearchResults = models.SearchResults{}
		a.state.LoadingSearchResults = false
		return nil
	}

	query := a.state.SearchQuery
	a.state.LoadingSearchResults = true

	return tea.Cmd(func() tea.Msg {
		// Add a small delay to allow for more typing (debounce)
		time.Sleep(300 * time.Millisecond)
		
		// Check if query has changed (debounce logic)
		if query != a.state.SearchQuery {
			return SearchResult{Results: models.SearchResults{}, Error: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Limit to 5 results per section for initial search
		resp, err := a.navidromeClient.Search(ctx, query, 5, 5, 5) // 5 artists, 5 albums, 5 tracks
		if err != nil {
			return SearchResult{Error: err}
		}

		// Convert Navidrome search results to our models
		results := models.SearchResults{
			Artists: make([]models.Artist, len(resp.SubsonicResponse.SearchResult3.Artist)),
			Albums:  make([]models.Album, len(resp.SubsonicResponse.SearchResult3.Album)),
			Tracks:  make([]models.Track, len(resp.SubsonicResponse.SearchResult3.Song)),
		}

		// Convert artists
		for i, artist := range resp.SubsonicResponse.SearchResult3.Artist {
			results.Artists[i] = models.Artist{
				ID:         artist.ID,
				Name:       artist.Name,
				AlbumCount: artist.AlbumCount,
				StarredAt:  artist.Starred,
			}
		}

		// Convert albums
		for i, album := range resp.SubsonicResponse.SearchResult3.Album {
			results.Albums[i] = models.Album{
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

		// Convert tracks
		for i, song := range resp.SubsonicResponse.SearchResult3.Song {
			results.Tracks[i] = models.Track{
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

		return SearchResult{Results: results, Error: nil}
	})
}

// handleSearchSelection handles when a search result is selected
func (a *App) handleSearchSelection(queueOnly bool) (tea.Model, tea.Cmd) {
	totalArtists := len(a.state.SearchResults.Artists)
	totalAlbums := len(a.state.SearchResults.Albums)
	totalTracks := len(a.state.SearchResults.Tracks)
	
	selectedIndex := a.state.SelectedSearchIndex
	currentIndex := 0
	
	// Check artists section
	if selectedIndex < currentIndex+totalArtists {
		// Selected an artist - show artist modal (same behavior for both modes)
		artist := a.state.SearchResults.Artists[selectedIndex-currentIndex]
		a.state.ShowSearchModal = false
		return a, a.showArtistModal(artist)
	}
	currentIndex += totalArtists
	
	// Check artists MORE button
	if totalArtists == 5 && selectedIndex == currentIndex {
		return a, a.loadMoreSearchResults("artists")
	}
	if totalArtists == 5 {
		currentIndex++
	}
	
	// Check albums section
	if selectedIndex < currentIndex+totalAlbums {
		// Selected an album - show album modal (same behavior for both modes)
		albumIndex := selectedIndex - currentIndex
		album := a.state.SearchResults.Albums[albumIndex]
		a.state.ShowSearchModal = false
		return a, a.showAlbumModal(album)
	}
	currentIndex += totalAlbums
	
	// Check albums MORE button
	if totalAlbums == 5 && selectedIndex == currentIndex {
		return a, a.loadMoreSearchResults("albums")
	}
	if totalAlbums == 5 {
		currentIndex++
	}
	
	// Check tracks section
	if selectedIndex < currentIndex+totalTracks {
		// Selected a track - different behavior based on mode
		trackIndex := selectedIndex - currentIndex
		if trackIndex < len(a.state.SearchResults.Tracks) {
			track := a.state.SearchResults.Tracks[trackIndex]
			a.state.ShowSearchModal = false
			
			if queueOnly {
				// Queue only - just add to queue
				return a, a.addTrackToQueue(track)
			} else {
				// Play and queue remaining tracks from search results
				remainingTracks := a.state.SearchResults.Tracks[trackIndex:]
				
				if a.audioManager != nil {
					// Clear current queue and add the track selection
					a.audioManager.ClearQueue()
					a.audioManager.AddTracksToQueue(remainingTracks)
					// Start playing the first track (selected one)
					a.audioManager.PlayTrackAtIndex(0)
					
					// Log the action for user feedback
					a.logMessage(fmt.Sprintf("Playing: %s - %s (%d tracks queued from search)", 
						track.Artist, track.Title, len(remainingTracks)))
				} else {
					// Fallback if audio manager not available
					a.state.Queue = remainingTracks
					a.state.CurrentTrack = &track
					a.state.IsPlaying = true
					
					a.logMessage(fmt.Sprintf("Playing: %s - %s", track.Artist, track.Title))
				}
				
				return a, nil
			}
		}
	}
	currentIndex += totalTracks
	
	// Check tracks MORE button
	if totalTracks == 5 && selectedIndex == currentIndex {
		return a, a.loadMoreSearchResults("tracks")
	}
	
	return a, nil
}

// SearchMoreResult represents the result of loading more search results
type SearchMoreResult struct {
	Section string
	Artists []models.Artist
	Albums  []models.Album
	Tracks  []models.Track
	Error   error
}

// loadMoreSearchResults loads the next 5 results for the specified section
func (a *App) loadMoreSearchResults(section string) tea.Cmd {
	if a.navidromeClient == nil || len(a.state.SearchQuery) == 0 {
		return nil
	}

	query := a.state.SearchQuery
	a.state.LoadingSearchResults = true

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()


		// For now, we'll simulate offset by searching with higher limits and taking only the new results
		// This is a simplification - ideally the Navidrome client would support offset parameters
		totalArtistsNeeded := len(a.state.SearchResults.Artists)
		totalAlbumsNeeded := len(a.state.SearchResults.Albums) 
		totalTracksNeeded := len(a.state.SearchResults.Tracks)
		
		if section == "artists" {
			totalArtistsNeeded += 5
		}
		if section == "albums" {
			totalAlbumsNeeded += 5
		}
		if section == "tracks" {
			totalTracksNeeded += 5
		}

		resp, err := a.navidromeClient.Search(ctx, query, totalArtistsNeeded, totalAlbumsNeeded, totalTracksNeeded)
		if err != nil {
			return SearchMoreResult{Section: section, Error: err}
		}

		// Extract only the new results
		var newArtists []models.Artist
		var newAlbums []models.Album
		var newTracks []models.Track

		switch section {
		case "artists":
			// Get artists beyond what we already have
			startIdx := len(a.state.SearchResults.Artists)
			if len(resp.SubsonicResponse.SearchResult3.Artist) > startIdx {
				for i := startIdx; i < len(resp.SubsonicResponse.SearchResult3.Artist); i++ {
					artist := resp.SubsonicResponse.SearchResult3.Artist[i]
					newArtists = append(newArtists, models.Artist{
						ID:         artist.ID,
						Name:       artist.Name,
						AlbumCount: artist.AlbumCount,
						StarredAt:  artist.Starred,
					})
				}
			}
		case "albums":
			// Get albums beyond what we already have
			startIdx := len(a.state.SearchResults.Albums)
			if len(resp.SubsonicResponse.SearchResult3.Album) > startIdx {
				for i := startIdx; i < len(resp.SubsonicResponse.SearchResult3.Album); i++ {
					album := resp.SubsonicResponse.SearchResult3.Album[i]
					newAlbums = append(newAlbums, models.Album{
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
					})
				}
			}
		case "tracks":
			// Get tracks beyond what we already have
			startIdx := len(a.state.SearchResults.Tracks)
			if len(resp.SubsonicResponse.SearchResult3.Song) > startIdx {
				for i := startIdx; i < len(resp.SubsonicResponse.SearchResult3.Song); i++ {
					song := resp.SubsonicResponse.SearchResult3.Song[i]
					newTracks = append(newTracks, models.Track{
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
					})
				}
			}
		}

		return SearchMoreResult{
			Section: section,
			Artists: newArtists,
			Albums:  newAlbums,
			Tracks:  newTracks,
			Error:   nil,
		}
	})
}

// sortAlbumsAsync sorts albums using Navidrome API calls for accurate sorting
func (a *App) sortAlbumsAsync(sortBy string) tea.Cmd {
	if a.navidromeClient == nil {
		return nil
	}

	a.state.LoadingAlbums = true
	a.state.LoadingError = ""

	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var albumType string
		switch sortBy {
		case "alpha":
			albumType = "alphabeticalByName"
		case "album_artist":
			albumType = "alphabeticalByArtist"
		case "date_added":
			albumType = "newest"
		case "play_count":
			// Play count sorting filters to only played albums using "frequent", so use in-memory sorting instead
			return AlbumsSortResult{SortBy: sortBy, UseInMemorySort: true}
		case "year":
			// Year sorting not directly supported by API, fallback to in-memory
			return AlbumsSortResult{SortBy: sortBy, UseInMemorySort: true}
		default:
			albumType = "alphabeticalByName"
		}

		// Load ALL albums for sorting
		resp, err := a.navidromeClient.GetAlbumsByType(ctx, albumType, 10000, 0)
		if err != nil {
			return AlbumsSortResult{Error: err, SortBy: sortBy}
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
				PlayCount:  album.PlayCount,
				CreatedAt:  album.Created,
				CoverArt:   album.CoverArt,
			}
		}

		return AlbumsSortResult{Albums: albums, SortBy: sortBy}
	})
}

// AlbumsSortResult represents the result of an album sort operation
type AlbumsSortResult struct {
	Albums          []models.Album
	SortBy          string
	UseInMemorySort bool // Flag to indicate fallback to in-memory sorting
	Error           error
}

// sortAlbumsInMemory sorts albums in memory (fallback for API-unsupported sorts)
func (a *App) sortAlbumsInMemory(sortBy string) {
	albums := a.state.Albums
	switch sortBy {
	case "year":
		// Sort by year (descending - newest first)
		for i := 0; i < len(albums)-1; i++ {
			for j := 0; j < len(albums)-i-1; j++ {
				if albums[j].Year < albums[j+1].Year {
					albums[j], albums[j+1] = albums[j+1], albums[j]
				}
			}
		}
	case "play_count":
		// Sort by play count (descending - most played first)
		// This includes albums with 0 play count, unlike API "frequent" sort
		for i := 0; i < len(albums)-1; i++ {
			for j := 0; j < len(albums)-i-1; j++ {
				if albums[j].PlayCount < albums[j+1].PlayCount {
					albums[j], albums[j+1] = albums[j+1], albums[j]
				}
			}
		}
	// Add other fallback sorts if needed
	}
	
	// Reset selection to the beginning after sorting
	a.state.SelectedAlbumIndex = 0
}

// sortArtistsAsync sorts artists using in-memory sorting (API doesn't have great artist sorting)
func (a *App) sortArtistsAsync(sortBy string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// For artists, we'll use in-memory sorting since API doesn't provide good sorting options
		return ArtistsSortResult{SortBy: sortBy, UseInMemorySort: true}
	})
}

// ArtistsSortResult represents the result of an artist sort operation
type ArtistsSortResult struct {
	SortBy          string
	UseInMemorySort bool
	Error           error
}

// sortArtistsInMemory sorts artists in memory
func (a *App) sortArtistsInMemory(sortBy string) {
	artists := a.state.Artists
	switch sortBy {
	case "alpha":
		// Sort alphabetically by artist name
		for i := 0; i < len(artists)-1; i++ {
			for j := 0; j < len(artists)-i-1; j++ {
				if artists[j].Name > artists[j+1].Name {
					artists[j], artists[j+1] = artists[j+1], artists[j]
				}
			}
		}
	case "play_count":
		// Sort by play count (descending - most played first)
		for i := 0; i < len(artists)-1; i++ {
			for j := 0; j < len(artists)-i-1; j++ {
				if artists[j].PlayCount < artists[j+1].PlayCount {
					artists[j], artists[j+1] = artists[j+1], artists[j]
				}
			}
		}
	case "date_added":
		// For artists, sort by album count as a proxy for date added
		for i := 0; i < len(artists)-1; i++ {
			for j := 0; j < len(artists)-i-1; j++ {
				if artists[j].AlbumCount < artists[j+1].AlbumCount {
					artists[j], artists[j+1] = artists[j+1], artists[j]
				}
			}
		}
	}
	
	// Reset selection to the beginning after sorting
	a.state.SelectedArtistIndex = 0
}

// sortPlaylistsAsync sorts playlists using in-memory sorting
func (a *App) sortPlaylistsAsync(sortBy string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// For playlists, we'll use in-memory sorting since playlists tab is not fully implemented
		return PlaylistsSortResult{SortBy: sortBy, UseInMemorySort: true}
	})
}

// PlaylistsSortResult represents the result of a playlist sort operation
type PlaylistsSortResult struct {
	SortBy          string
	UseInMemorySort bool
	Error           error
}

// sortPlaylistsInMemory sorts playlists in memory
func (a *App) sortPlaylistsInMemory(sortBy string) {
	playlists := a.state.Playlists
	switch sortBy {
	case "alpha":
		// Sort alphabetically by playlist name
		for i := 0; i < len(playlists)-1; i++ {
			for j := 0; j < len(playlists)-i-1; j++ {
				if playlists[j].Name > playlists[j+1].Name {
					playlists[j], playlists[j+1] = playlists[j+1], playlists[j]
				}
			}
		}
	case "date_added":
		// Sort by creation date (descending - newest first)
		for i := 0; i < len(playlists)-1; i++ {
			for j := 0; j < len(playlists)-i-1; j++ {
				if playlists[j].CreatedAt.Before(playlists[j+1].CreatedAt) {
					playlists[j], playlists[j+1] = playlists[j+1], playlists[j]
				}
			}
		}
	case "play_count":
		// Sort by song count as a proxy for activity level
		for i := 0; i < len(playlists)-1; i++ {
			for j := 0; j < len(playlists)-i-1; j++ {
				if playlists[j].SongCount < playlists[j+1].SongCount {
					playlists[j], playlists[j+1] = playlists[j+1], playlists[j]
				}
			}
		}
	}
	
	// Reset selection to the beginning after sorting  
	// Note: Playlists might not have a selection index yet, this will be added when Playlists tab is implemented
}
