# Navitone-CLI

A modern terminal-based music player for Navidrome with intuitive UI, comprehensive scrobbling support, and robust MPV-powered audio playback.

## 🎵 Overview

Navitone-CLI brings the convenience of a graphical music player to the terminal, designed with the philosophy that terminal interfaces should be as intuitive as their GUI counterparts. Think Spotify or Feishin, but optimized for terminal use with sensible keybindings and context-aware navigation.

## ✨ Key Features

- **Intuitive TUI Interface** - Tab-based navigation with visual styling
- **Navidrome Integration** - Full API support for streaming and library management  
- **MPV-Powered Audio** - Professional audio playback with universal format support
- **Smart Navigation** - Modal-based browsing with intuitive keybindings
- **Queue Management** - Complete playback controls with smart queue management
- **Scrobbling Support** - Last.fm and ListenBrainz integration with offline queuing

## 🚀 Current Status

### ✅ Fully Implemented
- **MPV Audio Backend** - Professional-grade audio with universal format support and perfect seeking
- **Core Architecture** - MVC pattern with clean separation of concerns  
- **Tab Navigation** - 7 tabs: Home, Albums, Artists, Tracks, Playlists, Queue, Config
- **Interactive Configuration** - Form-based config with field validation and connection testing
- **Albums Tab** - Live browsing, modal track views, Alt+Enter quick queuing
- **Artists Tab** - Nested navigation (Artist → Albums → Tracks) with smart queue integration
- **Tracks Tab** - Live track browsing with direct queue integration
- **Queue Management** - Complete playback controls: play/pause, next/prev, volume, seeking
- **Modal System** - Seamless navigation flow with context-aware controls
- **Enhanced Keybindings** - Intuitive shortcuts (Space, Alt+arrows, Shift+arrows) with no vim-style keys
- **Enhanced Global Search** - Shift+S modal search with intelligent result limiting, pagination, and dual-mode playback
- **Scrobbling System** - Full Last.fm and ListenBrainz support with Now Playing updates
- **Process Management** - Proper MPV lifecycle with graceful shutdown and cleanup

### 🏗️ In Development
- **Playlists Tab** - User playlist management and creation
- **Home Tab** - Proper curated sections (Recently Added/Played, Most Played, Random Albums)
- **Sorting Options** - Sort controls for Albums, Artists, Tracks tabs
- **Advanced Queue Features** - Reorder tracks, shuffle mode, repeat modes

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
- Live data from Navidrome server with pagination support
- Format: `[Year] Artist - Album (Track count)`
- ↑↓ navigation with visual selection highlighting
- **Enter** - Opens album tracks modal with detailed track listing
- **Alt+Enter/A** - Queue entire album immediately (bypass modal)
- **R** - Refresh albums list, maintains selection position
- **Modal Features**: Track-by-track navigation, play from any track, queue remainder

### 🎤 Artists  
- Live artist browsing from Navidrome with full metadata
- Format: `Artist Name (X albums)` with starred favorites (★)
- ↑↓ navigation with selection highlighting
- **Enter** - Opens artist albums modal showing all albums by artist
- **R** - Refresh artists list
- **Nested Navigation**: Artist → Albums → Tracks with seamless modal transitions
- **Album Modal Features**: Enter = view tracks, Alt+Enter/A = queue all albums

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
- **✅ Full Playback Controls** - Enter/Space to play, Ctrl+N/P for next/previous
- **✅ Real Audio Playback** - Streaming audio from Navidrome with format support
- Shows current playing track with ▶/⏸ indicators
- Volume control with Shift+Up/Down keys
- Reorder tracks (Planned)

### ⚙️ Config
- **Navidrome Settings** - Server URL, credentials with connection testing
- **Scrobbling Services** - Last.fm and ListenBrainz configuration with validation
- **Audio Settings** - Volume, device selection, buffer size
- **Interactive Forms** - Navigate with ↑↓, edit with Enter, save with F2

## 🎵 Audio System

### MPV-Powered Backend ✅
- **Universal Format Support** - MPV handles all audio formats natively (FLAC, MP3, OGG, WAV, AAC, M4A, etc.)
- **Professional Audio Pipeline** - Battle-tested MPV audio processing and decoding
- **Perfect Seeking** - Frame-accurate seeking in all compressed and uncompressed formats
- **Network Streaming** - Robust HTTP streaming with intelligent buffering and retry logic
- **Gapless Playback** - Seamless track transitions without audio dropouts
- **Advanced Features** - Built-in replay gain, audio filters, and crossfading support

### Audio Pipeline ✅
- **Streaming** - MPV handles direct HTTP streaming from Navidrome server with retry logic
- **Format Detection** - MPV automatically detects and decodes all audio formats
- **Cross-platform Audio** - MPV manages audio output across different systems (Pulse/ALSA/Pipewire/etc.)
- **Real-time Playback** - Professional buffering with configurable buffer sizes
- **JSON IPC Control** - Full control over playback via MPV's JSON IPC interface

### Playback Features ✅
- **Queue Management** - Add/remove tracks, clear queue, visual navigation
- **Playback Controls** - Play, pause, resume, stop, next, previous
- **Volume Control** - Adjustable volume levels (0-100%)
- **State Tracking** - Real-time playback position and duration
- **Event System** - Proper callbacks for UI updates and scrobbling
- **Error Handling** - Graceful fallbacks and error recovery

## 🔧 Requirements

- **Go 1.21+**
- **MPV Media Player** - Install via your package manager (`sudo apt install mpv`, `brew install mpv`, etc.)
- **Linux/macOS/Windows** - Cross-platform support via MPV
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
- **Shift+S** - Enhanced global search with intelligent pagination and dual-mode playback
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
   - Use ↑↓ to navigate, Enter to view tracks in modal
   - Alt+Enter or A to queue entire album immediately
   - In album modal: Enter to play track + queue remainder
   - Press R to refresh the list
2. Navigate to **Artists** tab - browse by artist
   - See album counts and starred favorites (★)
   - Enter to view artist's albums in modal
   - Navigate albums → Enter to view tracks → play from any track
   - Alt+Enter or A to queue all albums from artist
3. Navigate to **Tracks** tab - browse individual tracks
   - Random tracks from your library
   - Enter to add individual tracks to queue
4. Navigate to **Queue** tab - manage your playback queue
   - X/Del to remove tracks, C to clear all
   - **✅ Enter/Space to play tracks with real audio**
   - **✅ Alt+Left/Right for next/previous, Shift+Up/Down for volume**

### Enhanced Navigation & Playback ✅ WORKING
- **Global Playback**: Space (play/pause), Alt+Left/Right (previous/next track)
- **Enhanced Global Search**: Shift+S opens intelligent search modal with:
  - Smart result limiting (5 per section: Artists, Albums, Tracks)
  - "MORE" pagination options for browsing additional results
  - Dual-mode playback: Enter (play + queue remaining) vs Shift+Enter (queue only)
  - Real-time search with organized, categorized results
- **Volume Control**: Shift+Up/Down for volume adjustment
- **Seeking**: Left/Right arrow keys for 10-second scrubbing
- **Multi-format Support**: FLAC, MP3, OGG, WAV streaming with real-time playback
- **Smart Queue Management**: Play from any track, queue remainder automatically
- **Modal Navigation**: Seamless drilling down from artists → albums → tracks
- **Quick Actions**: Alt+Enter for immediate queuing, bypass confirmation modals

## ⚙️ Configuration

Configuration is stored in `~/.config/navitone-cli/config.toml`:

```toml
[navidrome]
server_url = \"https://your-navidrome-server.com\"
username = \"your-username\"
password = \"your-password\"
timeout = 30

[audio]
volume = 100
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
- **Conventional keybindings** - No vim-style navigation, intuitive shortcuts
- **Clean architecture** - MVC pattern with clear separation  
- **User-friendly** - Terminal interface should feel intuitive
- **Professional audio** - MPV backend for reliable, high-quality playback

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes following the existing patterns
4. Test thoroughly
5. Submit a pull request

## 🎯 Roadmap

### Phase 1 (Core Functionality) - ✅ COMPLETE
- [x] **MPV Audio Backend** - Complete replacement of custom decoders with professional MPV system
- [x] **Interactive Configuration** - Forms with validation and connection testing
- [x] **Navidrome Integration** - Full API support for streaming and library management
- [x] **Albums Tab** - Modal track views with enhanced navigation and quick queuing
- [x] **Artists Tab** - Nested navigation (Artist → Albums → Tracks) with smart integration
- [x] **Tracks Tab** - Live browsing with direct queue integration
- [x] **Queue Management** - Complete playback controls with volume and seeking
- [x] **Modal System** - Intuitive navigation flow with context-aware controls  
- [x] **Enhanced Keybindings** - Clean, conventional shortcuts without vim-style navigation
- [x] **Enhanced Global Search** - Shift+S modal with intelligent pagination, dual-mode playback, and smart result limiting
- [x] **Process Management** - Proper MPV lifecycle with graceful shutdown
- [x] **Scrobbling Support** - Last.fm and ListenBrainz with Now Playing updates

### Phase 2 (Remaining Features)
- [ ] **Playlists Tab** - User playlist management and creation
- [ ] **Home Tab Curation** - Proper Recently Played, Most Played, Random sections
- [ ] **Sorting Options** - Sort controls for Albums, Artists, Tracks tabs  
- [ ] **Advanced Queue Features** - Reorder tracks, shuffle mode, repeat modes

### Phase 3 (Advanced Features)
- [ ] Advanced mouse support
- [ ] Album art display (ASCII)
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

**Note**: Phase 1 core functionality is **complete**! The application now features a professional MPV-powered audio backend with universal format support, intuitive modal navigation, smart queue management, and clean keybindings. Ready for daily use with any Navidrome server.

**Latest Update**: ✅ **Enhanced Global Search Complete** - Upgraded Shift+S global search with intelligent result limiting (5 per section), "MORE" pagination for browsing additional results, and dual-mode playback controls (Enter: play + queue remaining, Shift+Enter: queue only). Features smart navigation that prevents getting stuck on last items and seamless integration with existing modal workflows.