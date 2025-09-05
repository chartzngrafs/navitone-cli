# Navitone-CLI Requirements

## Project Overview
Terminal-based Navidrome music player with graphical TUI interface, intuitive navigation, and robust functionality.

## Technical Stack
- **Language**: Go
- **TUI Framework**: Bubble Tea
- **Audio**: Go audio library with Pipewire support (beep or oto)
- **Target Platform**: Linux (primary/only)
- **Config Format**: TOML

## Audio Requirements
### Supported Formats
- FLAC
- MP3
- OGG
- WAV

### Audio System
- Pipewire integration for low-latency playback
- System audio integration (PulseAudio fallback if needed)

## User Interface

### Navigation
- Tab-based interface with conventional keyboard shortcuts
- No vim-like keybindings - standard shortcuts only (Ctrl+C, Ctrl+V, Tab, Enter, arrows)
- Mouse support (Phase 2, but build foundation in Phase 1 if feasible)

### Interface Sections (Tabs)
1. **Home** - Default startup view
   - Recently Added Albums
   - Recently Played Albums
   - Most Played Albums
   - Random Albums

2. **Albums** - Browse all albums with sorting options ✅
   - Live data from Navidrome server
   - Format: [Year] Artist - Album (Track count)
   - ↑↓ navigation with visual selection
   - Enter to add album to queue
   - R to refresh, pagination for large collections

3. **Artists** - Browse by album artists with sorting options ✅
   - Live artist browsing from Navidrome
   - Format: Artist Name (X albums)
   - Star indicators (★) for favorited artists
   - ↑↓ navigation with selection highlighting
   - Enter to add artist to queue

4. **Tracks** - Browse individual tracks with sorting options ✅
   - Live track browsing from Navidrome (random tracks) ✅
   - Format: Track#. Artist - Title (Album) [Duration] ✅
   - ↑↓ navigation with selection highlighting ✅
   - Enter to add track to queue ✅
   - Sorting options ❌
   - Search and filter capabilities ❌

5. **Playlists** - User playlists management ❌
   - Currently shows "Coming soon" placeholder ❌
   - Create, edit, delete playlists ❌
   - Browse user playlists ❌

6. **Queue** - Current playback queue ⚠️ (Partial)
   - Visual queue management with navigation ✅
   - Add tracks from browse tabs ✅
   - Remove individual tracks (X/Del) ✅
   - Clear entire queue (C) ✅
   - Shows current "playing" track with indicators ✅
   - Simulate play/pause (Enter/Space) ✅
   - Reorder tracks ❌
   - Actual audio playback ❌

7. **Config** - Configuration and setup
   - Navidrome server connection settings
   - Scrobbling service configuration (Last.fm, ListenBrainz)
   - Audio and UI preferences

### Key Features
- Context-aware help menu overlay
- Sorting options available across all browse views
- Queue management from any view (add tracks to queue)
- Intuitive keybindings that "just make sense"

## Navidrome Integration
- Server URL and credentials stored in config
- Streaming-focused (offline mode not required for Phase 1)
- Full API integration for browsing and playback

## Scrobbling Support
- **Last.fm Integration**
  - Username/password authentication
  - Real-time scrobbling of played tracks
  - "Now Playing" updates
  - Configurable scrobbling threshold (e.g., 50% of track played)

- **ListenBrainz Integration**
  - Token-based authentication
  - Submit listening data to ListenBrainz
  - Support for "playing now" submissions
  
- **Scrobbling Features**
  - Enable/disable per service
  - Retry mechanism for failed submissions
  - Offline queue for scrobbles when network unavailable
  - Configurable minimum play time before scrobbling

## Configuration
- TOML configuration file
- Store server connection details
- User preferences and keybinding customization

## Development Phases

### Phase 1 (Core Functionality)
- All interface tabs and navigation ✅
- Interactive configuration system with forms ✅
- Navidrome API integration with connection testing ✅
- Albums tab with live data browsing ✅
- Artists tab with live data browsing ✅
- Tracks tab with live data browsing ✅
- Basic queue management (add/remove/clear) ✅
- Context-aware help system ✅
- Comprehensive scrobbling support (Last.fm & ListenBrainz) ✅
- **Audio playback with format support** ❌
- **Playlists tab with playlist management** ❌ 
- **Home tab with proper curated content sections** ⚠️ (Partial)
- **Sorting options for browse tabs** ❌
- **Actual playback controls** ❌

### Phase 2 (Enhancements)
- Full mouse support refinement
- Advanced features (lyrics, etc.)
- Performance optimizations

## Architecture Considerations
- MVC pattern with separate views for each tab
- Component-based UI with reusable elements
- Centralized state management for playback and navigation
- Configurable keybinding system