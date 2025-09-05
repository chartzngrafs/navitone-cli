# Navitone-CLI Requirements

## Project Overview
Terminal-based Navidrome music player with graphical TUI interface, intuitive navigation, and robust functionality.

## Technical Stack
- **Language**: Go
- **TUI Framework**: Bubble Tea
- **Audio**: Go audio library with Pipewire support (beep or oto)
- **Target Platform**: Linux (primary/only)
- **Config Format**: TOML

## Audio Requirements ✅ COMPLETE
### Supported Formats ✅ COMPLETE
- ✅ FLAC (implemented with proper int32→int16 conversion)
- ✅ MP3 (implemented with native go-mp3 decoder)
- ✅ OGG (implemented with proper float32→int16 conversion)
- ✅ WAV (implemented with basic WAV support)

### Audio System ✅ COMPLETE
- ✅ Oto audio library integration for cross-platform playback
- ✅ Proper PCM audio pipeline with format conversion
- ✅ Audio streaming from Navidrome server
- ✅ Volume control and playback state management

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

2. **Albums** - Browse all albums with sorting options

3. **Artists** - Browse by album artists with sorting options

4. **Tracks** - Browse individual tracks with sorting options

5. **Playlists** - User playlists management

6. **Queue** - Current playback queue
   - Clear queue functionality
   - Reorder tracks
   - Remove individual tracks

### Key Features
- Context-aware help menu overlay
- Sorting options available across all browse views
- Queue management from any view (add tracks to queue)
- Intuitive keybindings that "just make sense"

## Navidrome Integration
- Server URL and credentials stored in config
- Streaming-focused (offline mode not required for Phase 1)
- Full API integration for browsing and playback

## Configuration
- TOML configuration file
- Store server connection details
- User preferences and keybinding customization

## Development Phases

### Phase 1 (Core Functionality)
- All interface tabs and navigation
- ✅ Audio playbook with format support (COMPLETE - MP3, FLAC, OGG, WAV)
- ✅ Audio encoding/decoding pipeline (COMPLETE - proper PCM conversion)
- Navidrome API integration
- ✅ Queue management (COMPLETE - add, remove, reorder, clear)
- Basic configuration system
- Context-aware help system

### Phase 2 (Enhancements)
- Full mouse support refinement
- Advanced features (scrobbling, lyrics, etc.)
- Performance optimizations

## Architecture Considerations
- MVC pattern with separate views for each tab
- Component-based UI with reusable elements
- Centralized state management for playback and navigation
- Configurable keybinding system