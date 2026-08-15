# Navitone-CLI

A modern terminal-based music player for Navidrome with intuitive UI, comprehensive scrobbling support, robust MPV-powered audio playback, and complete Omarchy 4.0 Quattro theme integration.

## 🎵 Overview

Navitone-CLI brings the convenience of a graphical music player to the terminal, designed with the philosophy that terminal interfaces should be as intuitive as their GUI counterparts. Think Spotify or Feishin, but optimized for terminal use with sensible keybindings and context-aware navigation.

## ✨ Key Features

- **Omarchy Quattro 24-Color Moving Gradient & Visual Flair** - Dynamic multi-color Omarchy gradient body borders (`Shift+G`), Base24 & 24-color Quattro palette support, animated timeline progress wave, ambient selection row shimmer, and sleek tab navigation header accent
- **In-App Vertical Cava Audio Visualizer** - Multi-bar Cava spectrum visualizer (`Shift+C`) styled with vertical bottom-to-top theme gradient interpolation derived from active Omarchy `cava_theme` or theme palettes
- **On-Demand Live Theme Reload** - Instant live theme refresh (`Shift+R`) that re-reads active Omarchy Quattro state (`~/.local/state/omarchy/`) and updates both UI components and Cava visualizer live
- **Enhanced Theming System** - Rich 8-color semantic theming with extended 24-color Quattro palette support for desktop-synchronized colors
- **Intuitive TUI Interface** - Tab-based navigation with comprehensive visual styling and themed components
- **Navidrome Integration** - Full API support for streaming, library management, and internet radio proxying
- **MPV-Powered Audio** - Professional audio playback with universal format support, dynamic bitrate monitoring, and fallback handling
- **Real-Time ICY Stream Metadata** - Automatic live artist and song title parsing for web radio streams (Nightride FM, SomaFM, Icecast/Shoutcast)
- **Live Bitrate Display** - Real-time streaming bitrate indicator positioned to the left of volume in control panel and Zen Mode (e.g. `320 kbps • Vol: 80%`)
- **Smart Navigation** - Modal-based browsing with intuitive keybindings and themed progress indicators
- **Queue Management** - Complete playback controls with themed status indicators and progress bars
- **Unified Control Panel** - Streamlined 3-line bottom bar unifying player controls, dynamic progress timeline, live activity ticker, and context-aware keybind hints
- **Zen Mode** - Full-screen colored ASCII album art display (`Shift+Z`) with track info, playing state, bitrate, and progress bar
- **ASCII Album Art** - Display album artwork as ASCII art in Albums and Artists tabs
- **Scrobbling Support** - Last.fm and ListenBrainz integration with offline queuing

## 🚀 Current Status

### ✅ Fully Implemented
- **Omarchy Quattro Integration** - Multi-path resolution supporting Omarchy Quattro state (`~/.local/state/omarchy/current/theme/colors.toml`, `theme.name`, `cava_theme`), Base24, and 24-color palettes (`color0`..`color23`)
- **Vertical Cava Visualizer** - Integrated Cava spectrum visualizer (`Shift+C`) with bottom-to-top vertical color gradient interpolation derived from Omarchy `cava_theme` or active palette
- **Live Hotkey Reload (`Shift+R`)** - Instant global hotkey (`R` or `Shift+R`) that re-reads Omarchy Quattro theme files and updates UI & Cava colors live
- **Omarchy Moving Gradient & Visual Flair** - Dynamic, rotating TrueColor gradient borders surrounding main container boxes across all tabs (`Shift+G`), multi-color progress wave, row selection shimmer, and redesigned header tab bar
- **Streamlined Control Panel** - Single rounded 3-line panel replacing 3 separate legacy boxes, freeing 5+ lines of screen real estate for content
- **Zen Mode** - Distraction-free full-screen colored album art mode (`Shift+Z`), showing colored artwork, track metadata, playing status, live bitrate, and progress bar while hiding UI chrome
- **Enhanced Theming System** - Rich 8-color theming with 40+ styled components, progress bars, status indicators, and optional Omarchy desktop theme synchronization
- **MPV Audio Backend** - Professional-grade audio with universal format support, dynamic bitrate reporting (`audio-bitrate`), and perfect seeking
- **Internet Radio Proxying & Fallback** - Navidrome `/rest/stream?id=...` radio proxying with direct stream fallback, automatic Nightride FM `.m4a` -> `.mp3` mount normalization, and async stream retry logic
- **Real-Time ICY Stream Metadata** - MPV `media-title` / `icy-title` property observation that dynamically updates artist and track title as songs change on live internet radio streams
- **Live Bitrate Display** - Real-time audio bitrate indicator formatted to the left of volume (`320 kbps • Vol: 80%`) in player control panel and Zen Mode
- **Core Architecture** - MVC pattern with clean separation of concerns
- **Tab Navigation** - 7 tabs: Home, Albums, Artists, Playlists, Radio, Queue, Config with themed styling
- **Internet Radio Tab** - Live browsing and streaming of internet radio stations built into Navidrome server (`Enter`/`Space` play stream, `A`/`Alt+Enter` queue, `Shift+R` refresh)
- **Interactive Configuration** - Form-based config with field validation, connection testing, and optional Vim keybindings mode
- **Albums Tab** - Live browsing, modal track views, Alt+Enter quick queuing with themed indicators
- **Artists Tab** - Nested navigation (Artist → Albums → Tracks) with smart queue integration
- **Playlists Tab** - Complete playlist management with modal navigation, track-by-track playback, and queue integration
- **Interactive Home Tab** - Enhanced with 4 interactive sections: Recently Added Albums, Top Artists, Most Played Albums, and Top Tracks with ↑↓ navigation and real play count data
- **Queue Management** - Complete playback controls: play/pause, next/prev, volume, seeking with themed progress bars
- **Modal System** - Seamless navigation flow with context-aware controls across Albums, Artists, and Playlists
- **Enhanced Keybindings** - Intuitive shortcuts (Space, Alt+arrows, Shift+arrows) with optional Vim navigation mode (`vim_mode = true`)
- **Enhanced Global Search** - / modal search with intelligent result limiting, pagination, and dual-mode playback
- **ASCII Album Art System** - Configurable ASCII artwork display with Navidrome + MusicBrainz fallback, intelligent caching, and responsive layout
- **Scrobbling System** - Server-side scrobbling via Navidrome (preferred) with optional client-side Last.fm/ListenBrainz and Now Playing updates
- **Process Management** - Proper MPV lifecycle with graceful shutdown and cleanup

## 🎮 Keybindings Reference

| Hotkey | Action |
|--------|--------|
| `Tab` / `Shift+Tab` | Cycle through navigation tabs (Home, Albums, Artists, Playlists, Radio, Queue, Config) |
| `Space` | Play / Pause audio playback |
| `Alt+Right` / `Alt+Left` | Next track / Previous track |
| `Shift+Up` / `Shift+Down` | Volume up / Volume down |
| `Right` / `Left` | Seek forward 10s / Seek backward 10s |
| `Shift+R` or `R` | **Live reload active Omarchy Quattro theme and Cava visualizer colors** |
| `Shift+C` or `C` | Toggle in-app vertical Cava audio visualizer |
| `Shift+G` or `G` | Toggle animated Omarchy moving gradient borders and visual flair |
| `Shift+Z` or `Z` | Toggle Zen Mode (full-screen colored album art display) |
| `/` | Open global search modal |
| `Ctrl+C` or `q` | Quit application |

## ⚙️ Configuration

Configuration is stored in `~/.config/navitone-cli/config.toml`:

```toml
[navidrome]
server_url = "https://your-navidrome-server.com"
username = "your-username"
password = "your-password"
timeout = 30

[audio]
volume = 100
device = ""  # Auto-detect
buffer_size = 4096

[scrobbling]
method = "auto"

[ui]
theme = "dark"
gradient_border = true    # Enable Omarchy moving gradient & visual flair (Shift+G toggle)
cava_visualizer = true    # Enable integrated Cava spectrum visualizer by default (Shift+C toggle)
show_album_art = true     # Enable ASCII artwork display
vim_mode = false          # Enable Vim keybindings (j/k list, h/l tabs, g/G jump)
artwork_quality = "high" # Quality: low, medium, high, ultra
artwork_color = false     # Enable colored ASCII art
artwork_size = "medium"  # Size: small, medium, large
home_album_count = 8
accent_index = -1

[theme]  # Automatically populated by theme sync or Omarchy Quattro loader
name = "omarchy-firesky"
source = "omarchy"
background = "#0f0f0f"
foreground = "#e0d8d0"

[theme.colors]
accent = "#2a7a80"     # Headers, highlights
primary = "#2a7a80"    # Tabs, borders
secondary = "#c4753d"  # Selections, focus
success = "#1a6b6e"    # Play states, connections
warning = "#e68040"    # Loading, partial states
error = "#e6441a"      # Errors, disconnections
palette = ["#1a2424", "#e6441a", "#1a6b6e", "#e68040", "#2a7a80", "#c4753d", "#40b8b4", "#d8d0c8", "#4d6b6b", "#e66b40", "#40c4c0", "#e6a052", "#52c9cd", "#e6b375", "#70e0dd", "#f0e8e0"]
cava_palette = ["#e6441a", "#1a6b6e", "#e68040", "#2a7a80", "#c4753d", "#40b8b4"]
```

## 🎨 Omarchy Quattro Integration

Navitone-CLI includes native support for **Omarchy 4.0 Quattro**:
- **Quattro State Directories**: Resolves active runtime state under `~/.local/state/omarchy/current/theme/colors.toml`, `theme.name`, and `cava_theme`.
- **Base24 & 24-Color Schemas**: Supports line-by-line TOML parsing for `color0`..`color23` and `base00`..`base17`.
- **Vertical Cava Gradient**: Derives Cava spectrum colors from `cava_theme` or active palette (filtering background colors for maximum contrast).
- **Background Daemon**: `theme-sync/monitor-theme.sh` watches filesystem events and keeps `config.toml` updated automatically.

## 📦 Installation

```bash
git clone https://github.com/yourusername/navitone-cli.git
cd navitone-cli
make build && go install ./cmd/navitone
navitone
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
