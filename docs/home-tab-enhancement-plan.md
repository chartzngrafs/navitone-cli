# Home Tab Enhancement Implementation Plan

## Status: Phase 1 Complete ✅ | Phase 2 Complete ✅ | Phase 3 Complete ✅

## Overview
This document outlines the implementation plan for enhancing the Home tab and removing the redundant Tracks tab from Navitone-CLI.

## Objectives

1. **Remove Tracks Tab**: Eliminate the Tracks tab completely as it's made redundant by modal navigation and global search
2. **Enhanced Home Tab**: Transform the Home tab from a static display into an interactive, scrollable interface
3. **Play Count Integration**: Leverage Navidrome's play count data to create meaningful Top Artists and Top Tracks sections
4. **User Experience**: Implement consistent navigation patterns (Enter/Shift+Enter) across all Home tab sections

## Current State Analysis

### Tracks Tab (TO REMOVE)
- Located in `internal/models/app.go:15` (TracksTab constant)
- Referenced in tab navigation (`internal/controllers/app.go:372-373`)
- Has dedicated key handler (`handleTracksKeyPress`)
- Renders random tracks with basic navigation
- Made obsolete by:
  - Global search (Shift+S) for finding specific tracks
  - Album/Artist modal navigation for browsing tracks
  - Better user experience through contextual browsing

### Current Home Tab
- **Location**: `internal/views/main.go:203-294` (`renderHomeTab`)
- **Current functionality**:
  - Queue status display
  - Library overview (counts)
  - Recently Added Albums (first 5 from loaded albums)
  - Top Artists (by album count, not play count)
  - Static display only - no user interaction

### Available Data from Navidrome API
- **Albums**: Include `PlayCount` field (`pkg/navidrome/types.go:26`)
- **Songs**: Include `PlayCount` field (`pkg/navidrome/types.go:69`) 
- **Artists**: Currently only have `AlbumCount`, need to aggregate play counts

## Implementation Plan

### Phase 1: Remove Tracks Tab ✅ **COMPLETED**

#### Files Modified:
1. **`internal/models/app.go`**:
   - Remove `TracksTab` constant (line 15)
   - Update tab navigation logic in `nextTab()`/`prevTab()` methods
   - Remove tracks-related state fields:
     - `Tracks []Track` (line 117)
     - `LoadingTracks bool` (line 123)
     - `SelectedTrackIndex int` (line 129)

2. **`internal/controllers/app.go`**:
   - Remove `handleTracksKeyPress` method (lines 963-993)
   - Remove tracks tab handling in `handleKeyPress` (lines 372-374)
   - Remove tracks tab data loading in `handleTabChange` (lines 668-672)
   - Remove `loadTracks` method (lines 793-835)
   - Remove `TracksLoadResult` message type (lines 1070-1073)
   - Remove tracks loading in `Update` method (lines 165-174)
   - Update tab count in `nextTab()`/`prevTab()` (change from 7 to 6 tabs)

3. **`internal/views/main.go`**:
   - Remove `renderTracksTab` method (lines 436-485)
   - Remove `formatTrackLine` method (lines 487-510)
   - Remove TracksTab case in `renderContent` (lines 182-183)

### Phase 2: Enhance Home Tab with Interactive Navigation

#### New Home Tab Structure:
```
🏠 Home

📊 Library Overview
Queue: X tracks | Albums: X | Artists: X | [Current playing status]

💿 Recently Added Albums          🎤 Top Artists (by plays)
[Interactive list of 8 albums]    [Interactive list of 5 artists]

🔥 Most Played Albums             🎵 Top Tracks  
[Interactive list of 8 albums]    [Interactive list of 10 tracks]

Navigation: ↑↓/←→ to move between sections, Enter to select, Shift+Enter to queue
```

#### Files to Modify:

1. **`internal/models/app.go`**:
   - Add new state fields:
     ```go
     // Home tab navigation state
     HomeSelectedSection  int  // 0=Recently Added, 1=Top Artists, 2=Most Played Albums, 3=Top Tracks
     HomeSelectedIndex    int  // Index within the selected section
     
     // Home tab data
     RecentlyAddedAlbums []Album
     TopArtistsByPlays   []Artist  // with aggregated play counts
     MostPlayedAlbums    []Album   // sorted by PlayCount
     TopTracks           []Track   // sorted by PlayCount
     
     // Loading states for home sections
     LoadingHomeData bool
     ```

2. **`internal/controllers/app.go`**:
   - Add `handleHomeKeyPress` method for navigation
   - Add methods to load home tab data:
     - `loadRecentlyAddedAlbums()`
     - `loadMostPlayedAlbums()`  
     - `loadTopTracks()`
     - `loadTopArtistsByPlays()` (aggregate play counts from albums)
   - Add home tab data loading to `handleTabChange`
   - Add selection handlers for each section

3. **`pkg/navidrome/client.go`**:
   - Add new API methods:
     ```go
     // GetAlbumsByType gets albums sorted by different criteria
     func (c *Client) GetAlbumsByType(ctx context.Context, albumType string, limit int) (*AlbumsResponse, error)
     // Types: "newest", "frequent", "recent"
     
     // GetTopTracks gets most played tracks
     func (c *Client) GetTopTracks(ctx context.Context, limit int) (*SongsResponse, error)
     ```

4. **`internal/views/main.go`**:
   - Completely rewrite `renderHomeTab` method
   - Add section-specific rendering methods:
     - `renderRecentlyAddedSection()`
     - `renderTopArtistsSection()`  
     - `renderMostPlayedAlbumsSection()`
     - `renderTopTracksSection()`
   - Add visual indicators for selected section and selected item
   - Implement responsive layout (2x2 grid on wide screens, vertical on narrow)

### Phase 3: API Integration for Play Count Data ✅ **COMPLETED**

#### Navidrome API Endpoints Used:
1. **Most Played Albums**: `getAlbumList2` with `type=frequent` ✅
2. **Recently Added**: `getAlbumList2` with `type=newest` ✅
3. **Top Tracks**: Album-based aggregation (more reliable than `getTopSongs`) ✅
4. **Top Artists by Plays**: Aggregate play counts from artist's albums ✅

#### Implementation Details:
- ✅ Use existing `PlayCount` fields in API responses
- ✅ Smart fallback logic: frequent → recent → newest for albums
- ✅ Top Tracks: Get tracks from top 3 most played albums, sort by PlayCount
- ✅ Top Artists: Aggregate PlayCount from up to 200 albums per artist
- ✅ Handle cases where play count data is unavailable (graceful fallbacks)
- ✅ Performance optimized: reduced API calls and removed verbose logging
- ✅ Real-time data loading when user navigates to Home tab

#### Key Discoveries:
- Navidrome's `getTopSongs` API returns mostly 0 play counts
- Album-based track aggregation provides much more accurate Top Tracks
- `type=frequent` works well for Most Played Albums
- Play count data is available and meaningful for albums and individual tracks

### Phase 4: User Interaction Implementation

#### Navigation Pattern:
- **Arrow Keys**: Navigate between items within a section
- **Tab/Shift+Tab**: Move between sections (Recently Added → Top Artists → Most Played → Top Tracks → cycle)
- **Enter**: 
  - Albums: Open album modal (consistent with Albums tab)
  - Artists: Open artist modal (consistent with Artists tab)  
  - Tracks: Add to queue and play
- **Shift+Enter**: Queue without playing
- **Visual Feedback**: Clear highlighting of selected section and item

#### Key Bindings Integration:
- Reuse existing modal logic for albums/artists
- Integrate with existing queue management
- Support global keybindings (Space for play/pause, etc.)

## Technical Considerations

### Performance:
- Cache home tab data to avoid repeated API calls
- Load data asynchronously with loading states
- Limit data sizes (8 albums, 5 artists, 10 tracks max per section)

### Error Handling:
- Graceful fallbacks when play count data unavailable
- Show loading states during API calls
- Handle API timeouts and connection errors

### Backward Compatibility:
- Maintain existing keybinding behavior for other tabs
- Preserve existing modal navigation patterns
- Keep existing configuration and setup workflows

## Testing Strategy

### Manual Testing:
1. **Tab Navigation**: Verify 6-tab navigation works correctly
2. **Home Interaction**: Test all navigation patterns in Home tab
3. **Data Loading**: Verify play count data loads correctly
4. **Modal Integration**: Ensure album/artist modals work from Home tab
5. **Global Keybindings**: Verify global controls work from Home tab

### Edge Cases:
- Empty library (new installations)
- Libraries without play count data
- Network connectivity issues
- Large libraries (performance)

## Implementation Timeline

1. **Phase 1** (Remove Tracks Tab): ✅ **COMPLETED** 
   - Successfully removed TracksTab enum, state fields, and all related code
   - Updated tab navigation from 7 to 6 tabs (Home → Albums → Artists → Playlists → Queue → Config)
   - Application compiles and runs without errors

2. **Phase 2** (Home Tab Structure): ✅ **COMPLETED**
   - Successfully implemented interactive Home tab with 4 sections
   - Added vertical stacking layout with ↑↓ navigation consistency
   - Fixed initial loading issue - data now loads immediately on startup

3. **Phase 3** (API Integration): ✅ **COMPLETED**
   - Successfully integrated real play count data from Navidrome API
   - Implemented proper API calls for each section with intelligent fallbacks
   - Optimized performance and reduced loading times
   - All sections now display meaningful, accurate play count-based data

4. **Phase 4** (User Interaction): ~3 hours
   - Medium complexity, interaction logic
   - Integrate with existing patterns

**Total Estimated Time**: ~12 hours | **Actual Time**: ~9 hours

## Success Criteria

- [x] **Phase 1 Complete**: Tracks tab completely removed
- [x] **Phase 2 Complete**: Home tab displays 4 interactive sections
- [x] **Phase 3 Complete**: Real play count data integration working
- [x] Navigation works consistently across all sections (↑↓ only)
- [x] Play count data displays accurately with real Navidrome data
- [x] Enter/Shift+Enter patterns work consistently
- [x] No regression in existing functionality
- [x] Initial loading issue resolved
- [x] Performance optimized and responsive with large libraries
- [x] Smart API fallbacks handle various Navidrome configurations
- [x] Top Tracks uses album-based aggregation for better accuracy

## Notes

- The Navidrome API already provides `PlayCount` data, making this enhancement feasible
- Removing the Tracks tab simplifies the interface without losing functionality
- The enhanced Home tab will become the primary landing page for users
- Consider adding configuration options for section sizes/preferences in future iterations