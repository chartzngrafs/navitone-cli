# Navitone-CLI

A modern terminal-based music player for Navidrome with intuitive UI, comprehensive scrobbling support, and robust audio playback capabilities.

## 🎵 Overview

Navitone-CLI brings the convenience of a graphical music player to the terminal, designed with the philosophy that terminal interfaces should be as intuitive as their GUI counterparts. Think Spotify or Feishin, but optimized for terminal use with sensible keybindings and a context-aware help system.

## ✨ Key Features

- **Intuitive TUI Interface** - Tab-based navigation with visual styling
- **Navidrome Integration** - Full API support for streaming and library management
- **Comprehensive Scrobbling** - Last.fm and ListenBrainz support with offline queuing
- **Smart Navigation** - Context-aware help system and logical keybindings
- **Audio Excellence** - Pipewire support with multiple format handling (FLAC, MP3, OGG, WAV)
- **Mouse Support** - Full mouse interaction (Phase 2)

## 🚀 Current Status

### ✅ Implemented
- **Core Architecture** - MVC pattern with clean separation of concerns
- **Tab Navigation** - 7 tabs: Home, Albums, Artists, Tracks, Playlists, Queue, Config
- **Interactive Configuration** - Full form-based config with field validation
- **Navidrome API Client** - Complete Subsonic API integration with authentication
- **Scrobbling System** - Full Last.fm and ListenBrainz support with retry queuing
- **Albums Tab** - Live data browsing with navigation, selection, and queue integration
- **Artists Tab** - Live artist browsing with album counts and starred indicators
- **Tracks Tab** - Live track browsing with formatted display and queue integration
- **Basic Queue Management** - Add/remove tracks, clear queue, visual management
- **UI Framework** - Bubble Tea with Lipgloss styling and visual feedback
- **Help System** - Context-aware overlay with F1/? toggle
- **Connection Testing** - Async Navidrome server validation
- **Loading States** - Async data loading with error handling and retry

### 🏗️ In Development
- **Audio Playback System** - Pipewire integration with format support (FLAC, MP3, OGG, WAV)
- **Playlists Tab** - User playlist management and creation
- **Home Tab** - Proper curated sections (Recently Added/Played, Most Played, Random Albums)
- **Sorting Options** - Sort controls for Albums, Artists, Tracks tabs
- **Actual Playback Controls** - Real play/pause/next/previous functionality

### 📋 Planned (Phase 2)
- Enhanced mouse support
- Album art display (ASCII)
- Advanced features (lyrics, advanced queue management)
- Performance optimizations
- Playlist creation and management

## 🏗️ Architecture

```
navitone-cli/
├── cmd/navitone/           # Application entry point
├── internal/
│   ├── models/            # Data structures and types
│   ├── views/             # UI rendering and styling
│   ├── controllers/       # Business logic and event handling
│   ├── config/            # Configuration management
│   ├── audio/             # Audio playback system
│   └── api/              # Navidrome API client
├── pkg/
│   ├── navidrome/         # Reusable Navidrome client library
│   └── scrobbling/        # Last.fm & ListenBrainz scrobbling clients
└── docs/                  # Documentation and planning
```

## 🎛️ Interface Sections

### 🏠 Home Tab
- Library Overview - Current stats and queue status
- Recently Added Albums (Partial - shows first albums, needs proper curation)
- Top Artists by album count
- Recently Played Albums (Not Implemented)
- Most Played Albums (Not Implemented)  
- Random Albums (Not Implemented)

### 💿 Albums
- Live data from Navidrome server
- Format: `[Year] Artist - Album (Track count)`
- ↑↓ navigation with visual selection
- Enter to add album to queue
- R to refresh, pagination for large collections

### 🎤 Artists  
- Live artist browsing from Navidrome
- Format: `Artist Name (X albums)`
- Star indicators (★) for favorited artists
- ↑↓ navigation with selection highlighting
- Enter to add artist to queue

### 🎵 Tracks
- Live track browsing from Navidrome (random tracks)
- Format: `Track#. Artist - Title (Album) [Duration]`
- ↑↓ navigation with selection highlighting
- Enter to add track to queue
- Sorting options (Not Implemented)
- Search and filter capabilities (Not Implemented)

### 📋 Playlists
- **Not Implemented** - Currently shows "Coming soon"
- User playlist management (Planned)
- Create, edit, delete playlists (Planned)

### 🔄 Queue
- Visual queue management with ↑↓ navigation
- Add tracks from Albums, Artists, Tracks tabs
- X/Del to remove individual tracks
- C to clear entire queue
- Enter/Space to simulate play (no actual audio playback)
- Shows current "playing" track with ▶/⏸ indicators
- Reorder tracks (Not Implemented)

### ⚙️ Config
- **Navidrome Settings** - Server URL, credentials with connection testing
- **Scrobbling Services** - Last.fm and ListenBrainz configuration with validation
- **Audio Settings** - Volume, device selection, buffer size
- **Interactive Forms** - Navigate with ↑↓, edit with Enter, save with F2

## 🔧 Requirements

- **Go 1.21+**
- **Linux** (primary target platform)
- **Pipewire** (recommended) or PulseAudio
- **Navidrome Server** (for music streaming)

## 📦 Installation

### From Source
```bash
git clone https://github.com/yourusername/navitone-cli.git
cd navitone-cli
go build -o bin/navitone ./cmd/navitone
./bin/navitone
```

### Dependencies
The application will automatically download required Go dependencies:
- Bubble Tea (TUI framework)
- Lipgloss (styling)
- TOML parser

## 🎮 Usage

### Basic Navigation
- **Tab/Shift+Tab** - Switch between tabs
- **F1 or ?** - Toggle context-sensitive help
- **Ctrl+C or q** - Quit application

### First Run Setup
1. Launch the application: `./bin/navitone`
2. Navigate to the **Config** tab (rightmost tab)
3. Use ↑↓ arrows to navigate fields, Enter to edit
4. Enter your Navidrome server details
5. Configure scrobbling services (optional - toggle checkboxes with Enter)
6. Press F2 to save settings
7. Press F3 to test Navidrome connection

### Browse Your Music Library
1. Navigate to **Albums** tab - browse your album collection
   - Use ↑↓ to navigate, Enter to add to queue
   - Press R to refresh the list
2. Navigate to **Artists** tab - browse by artist
   - See album counts and starred favorites (★)
   - Same navigation: ↑↓ and Enter to add to queue
3. Navigate to **Tracks** tab - browse individual tracks
   - Random tracks from your library
   - Enter to add individual tracks to queue
4. Navigate to **Queue** tab - manage your playback queue
   - X/Del to remove tracks, C to clear all
   - Enter/Space to simulate play (no actual audio)

### Playback Controls (Coming Soon)
- **Space** - Play/Pause
- **Ctrl+N** - Next track
- **Ctrl+P** - Previous track
- **+/-** - Volume control

## ⚙️ Configuration

Configuration is stored in `~/.config/navitone-cli/config.toml`:

```toml
[navidrome]
server_url = \"https://your-navidrome-server.com\"
username = \"your-username\"
password = \"your-password\"
timeout = 30

[audio]
volume = 70
device = \"\"  # Auto-detect
buffer_size = 4096

[scrobbling.lastfm]
enabled = false
username = \"\"
password = \"\"
api_key = \"\"
secret = \"\"

[scrobbling.listenbrainz]
enabled = false
token = \"\"

[ui]
theme = \"dark\"
show_album_art = false
home_album_count = 8
```

## 🧪 Development

### Setup
```bash
git clone https://github.com/yourusername/navitone-cli.git
cd navitone-cli
go mod tidy
```

### Build
```bash
go build -o bin/navitone ./cmd/navitone
```

### Project Principles
- **Conventional keybindings** - No vim-style navigation
- **Context-aware help** - Help changes based on current view
- **Clean architecture** - MVC pattern with clear separation
- **User-friendly** - Terminal interface should feel intuitive

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes following the existing patterns
4. Test thoroughly
5. Submit a pull request

## 🎯 Roadmap

### Phase 1 (Core Functionality)
- [x] Interactive configuration forms with validation
- [x] Navidrome API integration with connection testing
- [x] Comprehensive scrobbling support (Last.fm & ListenBrainz)
- [x] Albums tab with live data browsing and queue integration
- [x] Artists tab with live data browsing and selection
- [x] Tracks tab with live data browsing and queue integration
- [x] Basic queue management (add/remove/clear, visual display)
- [ ] **Audio playback system with Pipewire** (Critical Missing)
- [ ] **Playlists tab with playlist management** (Not Started)
- [ ] **Home tab with proper curated sections** (Partially Done)
- [ ] **Sorting options for browse tabs** (Not Implemented)
- [ ] **Actual playback controls** (Simulated Only)

### Phase 2 (Enhancements)
- [ ] Advanced mouse support
- [ ] Album art display
- [ ] Lyrics integration
- [ ] Performance optimizations
- [ ] Plugin system

## 🤝 Scrobbling Services

### Last.fm Setup
1. Create a Last.fm API application
2. Get your API key and secret
3. Enter credentials in Config tab
4. Enable scrobbling

### ListenBrainz Setup  
1. Create a ListenBrainz account
2. Generate a user token
3. Enter token in Config tab
4. Enable scrobbling

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Navidrome](https://github.com/navidrome/navidrome) - Excellent music server
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Powerful TUI framework
- [Feishin](https://github.com/jeffvli/feishin) - UI/UX inspiration

---

**Note**: This project is in active development. Many features are planned but not yet implemented. See the Current Status section above for what's currently working.