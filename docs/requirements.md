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

### Interface Sections (Tabs) ✅ ENHANCED - 6 Tabs Total
1. **Home** ✅ ENHANCED - Interactive music dashboard with real play count data
   - Library overview and queue status
   - Recently Added Albums (8 albums with play counts)
   - Top Artists (5 artists by aggregated play count)  
   - Most Played Albums (8 albums sorted by actual play count)
   - Top Tracks (10 tracks from most played albums)
   - ↑↓ navigation between sections, Enter/Shift+Enter actions

2. **Albums** ✅ - Browse all albums with enhanced modal navigation
   - Live data from Navidrome with pagination
   - Enter: View album tracks in modal
   - Alt+Enter/A: Queue entire album immediately
   - Modal: Play from any track + queue remainder

3. **Artists** ✅ - Browse by album artists with nested navigation
   - Live artist browsing with metadata
   - Enter: View artist albums in modal  
   - Modal navigation: Albums → Tracks with seamless transitions
   - Alt+Enter/A: Queue all albums from artist

4. **Tracks** ✅ REMOVED - Track browsing now available through modals and global search
   - Functionality moved to Home tab Top Tracks section
   - Global search (Shift+F) provides track discovery
   - Modal navigation from Albums/Artists provides track access

5. **Playlists** ✅ - User playlists management with full modal navigation
   - ✅ View all user playlists with track counts and metadata
   - ✅ Enter: View playlist tracks in modal with track-by-track navigation
   - ✅ Alt+Enter/A: Queue entire playlist immediately
   - ✅ Modal: Play from any track + queue remainder
   - ✅ PgUp/PgDn navigation support for large playlists
   - [ ] Create, edit, delete playlists (Planned for Phase 2)

6. **Queue** ✅ - Current playback queue with full controls
   - ✅ Visual queue management with navigation
   - ✅ Play from any track in queue
   - ✅ Remove individual tracks (X/Del)
   - ✅ Clear entire queue (C)
   - ✅ Real-time playback controls with MPV backend

7. **Config** ✅ - Interactive configuration with validation
   - ✅ Navidrome server settings with connection testing
   - ✅ Last.fm and ListenBrainz scrobbling configuration
   - ✅ Audio settings (volume, device, buffer size)
   - ✅ Form-based interface with field validation

### Key Features ✅ IMPLEMENTED
- ✅ Context-aware help menu overlay with comprehensive keybinding documentation
- ✅ Modal-based navigation system (albums → tracks, artists → albums → tracks)
- ✅ Enhanced keybindings (Alt+Enter quick actions, context-aware controls)
- ✅ Smart queue management (play from any track, queue remainder automatically)
- ✅ Intuitive navigation flow that "just makes sense"
- ✅ Enhanced Global Search (Shift+F) - intelligent pagination, dual-mode playback
- [ ] Sorting options available across all browse views (Planned - Shift+S)
- ✅ Multiple ways to add content to queue from any view

## Navidrome Integration
- Server URL and credentials stored in config
- Streaming-focused (offline mode not required for Phase 1)
- Full API integration for browsing and playback

## Configuration
- TOML configuration file
- Store server connection details
- User preferences and keybinding customization

## Development Phases

### Phase 1 (Core Functionality) ✅ COMPLETE
- ✅ All interface tabs and navigation with enhanced modal system
- ✅ MPV-powered audio backend (COMPLETE - universal format support)
- ✅ Professional audio pipeline via MPV (COMPLETE - streaming, seeking, gapless)
- ✅ Navidrome API integration (COMPLETE - full Subsonic API support)
- ✅ Queue management (COMPLETE - add, remove, clear, smart playback)
- ✅ Enhanced navigation system (COMPLETE - modal flows, Alt+Enter shortcuts)
- ✅ Configuration system (COMPLETE - interactive forms with validation)
- ✅ Context-aware help system (COMPLETE - comprehensive keybinding docs)

### Phase 2 (Enhancements)
- Full mouse support refinement
- Advanced features (scrobbling, lyrics, etc.)
- Performance optimizations

## Architecture Considerations
- MVC pattern with separate views for each tab
- Component-based UI with reusable elements
- Centralized state management for playback and navigation
- Configurable keybinding system