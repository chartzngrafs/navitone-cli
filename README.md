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
- **Audio Visualizer** - Integrated Cava support with Shift+C hotkey for new terminal launch
- **ASCII Album Art** - Display album artwork as ASCII art in Albums and Artists tabs
- **Scrobbling Support** - Last.fm and ListenBrainz integration with offline queuing

## 🚀 Current Status

### ✅ Fully Implemented
- **MPV Audio Backend** - Professional-grade audio with universal format support and perfect seeking
- **Core Architecture** - MVC pattern with clean separation of concerns  
- **Tab Navigation** - 6 tabs: Home, Albums, Artists, Playlists, Queue, Config
- **Interactive Configuration** - Form-based config with field validation and connection testing
- **Albums Tab** - Live browsing, modal track views, Alt+Enter quick queuing
- **Artists Tab** - Nested navigation (Artist → Albums → Tracks) with smart queue integration
- **Playlists Tab** - Complete playlist management with modal navigation, track-by-track playback, and queue integration
- **Interactive Home Tab** - Enhanced with 4 interactive sections: Recently Added Albums, Top Artists, Most Played Albums, and Top Tracks with ↑↓ navigation and real play count data
- **Queue Management** - Complete playback controls: play/pause, next/prev, volume, seeking
- **Modal System** - Seamless navigation flow with context-aware controls across Albums, Artists, and Playlists
- **Enhanced Keybindings** - Intuitive shortcuts (Space, Alt+arrows, Shift+arrows) with no vim-style keys
- **Enhanced Global Search** - Shift+F modal search with intelligent result limiting, pagination, and dual-mode playback
- **ASCII Album Art System** - Configurable ASCII artwork display with Navidrome + MusicBrainz fallback, intelligent caching, and responsive layout
- **Scrobbling System** - Server-side scrobbling via Navidrome (preferred) with optional client-side Last.fm/ListenBrainz and Now Playing updates
- **Process Management** - Proper MPV lifecycle with graceful shutdown and cleanup

### 🏗️ In Development
- **Sorting Options** - Sort controls for Albums, Artists, Playlists tabs  
- **Advanced Queue Features** - Reorder tracks, shuffle mode, repeat modes
- **Playlist Management** - Create, edit, delete user playlists

### 📋 Planned (Phase 2)
- Enhanced mouse support
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
- **Interactive Navigation** - ↑↓ arrows to navigate through all sections seamlessly
- **Recently Added Albums** - Latest 8 albums with real play counts, Enter to view tracks modal
- **Top Artists** - Top 5 artists by aggregated play count with Enter to view artist modal  
- **Most Played Albums** - 8 most frequently played albums sorted by actual play count with modal access
- **Top Tracks** - 10 most played individual tracks from top albums with Enter to play + queue remaining, Shift+Enter to queue only
- **Smart Integration** - All sections support Enter/Shift+Enter patterns consistent with other tabs
- **Real Data** - All sections display genuine play count data from your Navidrome server

### 💿 Albums
- Live data from Navidrome server with infinite pagination support
- Format: `[Year] Artist - Album (Track count)`
- ↑↓ navigation with visual selection highlighting
- **Enter** - Opens album tracks modal with detailed track listing
- **Alt+Enter/A** - Queue entire album immediately (bypass modal)
- **R** - Refresh albums list, maintains selection position
- **M** - Load more albums (loads next 50 when available)
- **Smart Pagination**: Shows "more available - press M to load" when additional albums exist
- **Modal Features**: Track-by-track navigation, play from any track, queue remainder

### 🎤 Artists  
- Live artist browsing from Navidrome with full metadata
- Format: `Artist Name (X albums)` with starred favorites (★)
- ↑↓ navigation with selection highlighting
- **Enter** - Opens artist albums modal showing all albums by artist
- **R** - Refresh artists list
- **Nested Navigation**: Artist → Albums → Tracks with seamless modal transitions
- **Album Modal Features**: Enter = view tracks, Alt+Enter/A = queue all albums

### 🎵 Track Access
- **Enhanced Home Tab** - Browse top tracks directly in Home tab with seamless navigation
- **Album Modals** - Access all tracks from any album with detailed track listings
- **Artist Modals** - Browse artist albums and access their tracks through nested navigation  
- **Global Search** - Shift+F to search and find specific tracks across your entire library

### 📋 Playlists
- Complete playlist management with intuitive navigation
- Format: `Playlist Name (X tracks) - Owner`
- ↑↓ navigation with selection highlighting and PgUp/PgDn support
- **Enter** - Opens playlist tracks modal with detailed track listing
- **Alt+Enter/A** - Queue entire playlist immediately (bypass modal)
- **R** - Refresh playlists list
- **Modal Features**: Track-by-track navigation, play from any track, queue remainder
- **Smart Integration**: Consistent Enter/Shift+Enter patterns with Albums and Artists tabs

### 🔄 Queue
- Visual queue management with ↑↓ navigation
- Add tracks from Albums, Artists, Playlists tabs
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

- **Go 1.23+**
- **MPV Media Player** - Install via your package manager (`sudo apt install mpv`, `brew install mpv`, etc.)
- **Cava Audio Visualizer** - Install via your package manager (`sudo apt install cava`, `brew install cava`, etc.)
- **Linux/macOS/Windows** - Cross-platform support via MPV
- **Navidrome Server** (for music streaming)

## 📦 Installation

### From Source
```bash
git clone https://github.com/yourusername/navitone-cli.git
cd navitone-cli
# Option A: Makefile helpers
make build   # builds to bin/navitone
./bin/navitone

# Option B: go directly
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
- **Shift+F** - Enhanced global search with intelligent pagination and dual-mode playback
- **Shift+C** - Launch Cava audio visualizer in new terminal window
- **Ctrl+C or q** - Quit application

### First Run Setup
1. Launch the application: `./bin/navitone`
2. Navigate to the **Config** tab (rightmost tab)
3. Use ↑↓ arrows to navigate fields, Enter to edit
4. Enter your Navidrome server details
5. Scrobbling: If your Navidrome admin linked Last.fm/ListenBrainz, server-side scrobbling works automatically. The Config tab shows a status line. Client-side setup is optional.
6. Press F2 to save settings
7. Press F3 to test Navidrome connection

### Browse Your Music Library
1. Navigate to **Home** tab - your music dashboard
   - Use ↑↓ to navigate through 4 curated sections
   - Enter to open modals (albums/artists) or play tracks
   - Shift+Enter to queue tracks without playing
   - Press R to refresh all home data
2. Navigate to **Albums** tab - browse your album collection
   - Use ↑↓ to navigate, Enter to view tracks in modal
   - Alt+Enter or A to queue entire album immediately
   - In album modal: Enter to play track + queue remainder
   - Press R to refresh the list
   - Press M to load more albums (loads next 50 when available)
3. Navigate to **Artists** tab - browse by artist
   - See album counts and starred favorites (★)
   - Enter to view artist's albums in modal
   - Navigate albums → Enter to view tracks → play from any track
   - Alt+Enter or A to queue all albums from artist
4. Navigate to **Playlists** tab - browse your user playlists
   - See all playlists with track counts and owner information
   - Enter to view playlist tracks in modal with navigation
   - Alt+Enter or A to queue entire playlist immediately
   - Modal: Play from any track + queue remainder automatically
5. Navigate to **Queue** tab - manage your playback queue
   - X/Del to remove tracks, C to clear all
   - **✅ Enter/Space to play tracks with real audio**
   - **✅ Alt+Left/Right for next/previous, Shift+Up/Down for volume**

### Enhanced Navigation & Playback ✅ WORKING
- **Global Playback**: Space (play/pause), Alt+Left/Right (previous/next track)
- **Enhanced Global Search**: Shift+F opens intelligent search modal with:
  - Smart result limiting (5 per section: Artists, Albums, Tracks)
  - "MORE" pagination options for browsing additional results
  - Dual-mode playback: Enter (play + queue remaining) vs Shift+Enter (queue only)
  - Real-time search with organized, categorized results
- **Audio Visualizer**: Shift+C launches Cava in new terminal window with cross-platform support
- **Volume Control**: Shift+Up/Down for volume adjustment
- **Seeking**: Left/Right arrow keys for 10-second scrubbing
- **Multi-format Support**: FLAC, MP3, OGG, WAV streaming with real-time playback
- **Smart Queue Management**: Play from any track, queue remainder automatically
- **Modal Navigation**: Seamless drilling down from artists → albums → tracks
- **Quick Actions**: Alt+Enter for immediate queuing, bypass confirmation modals

## 🎨 ASCII Album Art

### Features
- **Smart Display** - Shows ASCII artwork below album/artist listings when selected
- **Dual Sources** - Uses Navidrome cover art URLs first, falls back to MusicBrainz Cover Art Archive
- **Intelligent Caching** - Caches converted artwork locally to improve performance
- **Responsive Layout** - Automatically adjusts item count to make room for artwork
- **Config Toggle** - Enable/disable via Config tab "Show Artwork" checkbox

### How It Works
1. Navigate to Albums or Artists tab
2. Enable "Show Artwork" in Config tab
3. Select any album/artist - artwork appears below the list
4. First load may take a moment (downloads and converts image)
5. Subsequent views use cached ASCII art for instant display

### Technical Details
- **Quality Levels** - Low (10 chars), Medium (69 chars), High (optimized), Ultra (braille)
- **Resolution Options** - Small (35x18), Medium (50x25), Large (70x35)
- **Color Support** - Full 24-bit color for modern terminals
- **Cache Location** - `~/.cache/navitone-cli/artwork/`
- **Cache Expiration** - 30 days
- **Fallback Chain** - Navidrome → MusicBrainz Cover Art Archive → None
- **Format Support** - All formats supported by ascii-image-converter library

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

[scrobbling]
# Select scrobbling method: "auto", "server", "client", or "disabled"
method = \"auto\"

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
show_album_art = true     # Enable ASCII artwork display
artwork_quality = \"high\" # Quality: low, medium, high, ultra
artwork_color = false     # Enable colored ASCII art
artwork_size = \"medium\"  # Size: small, medium, large
home_album_count = 8
accent_index = -1
```

Notes:
- When `method = "auto"` (default), Navitone uses server-side scrobbling if available for your user on Navidrome, and falls back to client-side if not configured or fails.
- The Config tab displays a status line: “Server Scrobbling Enabled/Disabled” based on your Navidrome user profile.

### Scrobbling Modes
- `auto` (default): Try Navidrome server-side scrobbling; fall back to client services if needed.
- `server`: Force server-side scrobbling via Navidrome; do not fall back.
- `client`: Use Last.fm and/or ListenBrainz only (if enabled in config).
- `disabled`: Don’t scrobble.

### Theming
- Themes: `dark` (default) and `light`.
- Change via `[ui] theme = "dark" | "light"` in `config.toml`.
- Palette highlights: Blue (headers), Aquamarine (active), Plum (errors), tuned for readability.

## 🧪 Development

### Setup
```bash
git clone https://github.com/yourusername/navitone-cli.git
cd navitone-cli
go mod tidy
```

### Run (development)
```bash
make run   # or: go run ./cmd/navitone
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
- [x] **Queue Management** - Complete playback controls with volume and seeking
- [x] **Modal System** - Intuitive navigation flow with context-aware controls  
- [x] **Enhanced Keybindings** - Clean, conventional shortcuts without vim-style navigation
- [x] **Enhanced Global Search** - Shift+F modal with intelligent pagination, dual-mode playback, and smart result limiting
- [x] **Interactive Home Tab** - Enhanced with 4 curated sections, seamless ↑↓ navigation, and real play count data integration
- [x] **Process Management** - Proper MPV lifecycle with graceful shutdown
- [x] **Scrobbling Support** - Last.fm and ListenBrainz with Now Playing updates

### Phase 2 (Remaining Features)
- [x] **Playlists Tab** - User playlist viewing and playback (COMPLETE)
- [x] **ASCII Album Art System** - Configurable artwork display with caching and fallback sources (COMPLETE)
- [ ] **Playlist Management** - Create, edit, delete user playlists  
- [ ] **Sorting Options** - Sort controls for Albums, Artists, Playlists tabs
- [ ] **Advanced Queue Features** - Reorder tracks, shuffle mode, repeat modes

### Phase 3 (Advanced Features)
- [ ] Advanced mouse support
- [ ] Lyrics integration  
- [ ] Performance optimizations
- [ ] Plugin system

## 🤝 Scrobbling

### Recommended: Server-side via Navidrome
- Ask your Navidrome admin to configure Last.fm and/or ListenBrainz integrations on the server and link your account in the Navidrome web UI.
- Navitone will detect this and scrobble via the server automatically. No client setup needed.

### Optional: Client-side services
If the server isn’t configured or you prefer client-side:
- Last.fm: Create an API application, obtain API key and secret, and enter them along with your username/password in the Config tab; enable the checkbox.
- ListenBrainz: Generate a user token and enter it in the Config tab; enable the checkbox.

Tip: Leave `scrobbling.method = "auto"` to try server first, fallback to client.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🐛 Bug Fixes & Known Issues

### Navigation Header Disappearing (RESOLVED ✅)
**Issue**: Navigation header would disappear after playing tracks from modal windows (Albums/Artists modals).

**Root Cause**: The MPV audio backend's player module (`renderPlayer()`) was interfering with the UI render cycle during modal-to-playback state transitions. When tracks started playing from modals, rapid state updates caused the player module to render with inconsistent state, corrupting the overall layout.

**Technical Fix**: 
- Removed player module from main render sequence to eliminate interference
- Implemented comprehensive queue state synchronization via audio manager callbacks  
- Added defensive rendering protections for invalid window dimensions
- Fixed ANSI color code handling in queue display formatting

**Files Modified**:
- `internal/views/main.go`: Removed player module render, fixed queue formatting
- `internal/audio/mpv/manager.go`: Enhanced state change notifications
- `internal/controllers/app.go`: Improved queue synchronization, modal state management

**Impact**: Modal navigation and track playback now work seamlessly without UI corruption.

## 🙏 Acknowledgments

- [Navidrome](https://github.com/navidrome/navidrome) - Excellent music server
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Powerful TUI framework
- [Feishin](https://github.com/jeffvli/feishin) - UI/UX inspiration

---

**Note**: Phase 1 core functionality is **complete**! The application now features a professional MPV-powered audio backend with universal format support, intuitive modal navigation, smart queue management, and clean keybindings. Ready for daily use with any Navidrome server.

**Latest Update**: ✅ **ASCII Album Art System Implementation** - Fully implemented configurable ASCII artwork display system. Features include: smart artwork display below album listings when selected, dual source support (Navidrome + MusicBrainz fallback), intelligent local caching system, responsive layout that adjusts item count for artwork space, config toggles for enable/disable and quality settings, and automatic play count aggregation for artists. The artwork system provides high-quality ASCII art conversion with multiple resolution and quality options, enhancing the visual experience while maintaining terminal compatibility.
