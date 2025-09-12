# ASCII Artwork Feature Implementation

## Overview
Add ASCII artwork display functionality to Navitone-CLI for Albums and Artists tabs, showing cover art below the listings when items are selected.

## Features
1. **Config Toggle**: User-controllable artwork display via Config tab
2. **Dual Source**: Primary from Navidrome, fallback to MusicBrainz Cover Art Archive
3. **ASCII Conversion**: Convert images to terminal-friendly ASCII art
4. **Smart Layout**: Display artwork only when screen space permits
5. **Caching**: Avoid redundant conversions and API calls

## Technical Architecture

### 1. Configuration System
- **Location**: `internal/config/config.go`
- **Field**: `UIConfig.ShowAlbumArt` (already exists)
- **Form**: Add toggle to Config tab form

### 2. ASCII Conversion
- **Library**: `github.com/TheZoraiz/ascii-image-converter`
- **Features**: Color support, configurable dimensions, braille patterns
- **Location**: New package `internal/artwork/`

### 3. Data Sources
- **Primary**: Navidrome cover art URLs (already available in Album.CoverArt)
- **Fallback**: MusicBrainz Cover Art Archive API
  - Endpoint: `https://coverartarchive.org/release/{mbid}/front`
  - Size: 250px thumbnails for faster conversion

### 4. UI Integration
- **Target Tabs**: Albums and Artists
- **Display Logic**: Show ASCII art below listings for selected items
- **Space Management**: Reduce visible items when artwork displayed
- **Minimum Requirements**: contentHeight > 25 chars

## Implementation Plan

### Phase 1: Config System (30 minutes)
1. Add `ShowArtworkField` to `ConfigFormField` enum
2. Update form rendering and validation functions
3. Test config toggle functionality

### Phase 2: ASCII Conversion Core (2 hours)
1. Create `internal/artwork/converter.go`
2. Integrate ascii-image-converter library
3. Implement basic image-to-ASCII conversion
4. Add configuration options (size, characters, colors)

### Phase 3: UI Layout Integration (4 hours)
1. Modify `internal/views/main.go` Albums/Artists rendering
2. Add artwork display area calculation
3. Implement smart viewport adjustment
4. Add artwork rendering in selected item context

### Phase 4: MusicBrainz Integration (3 hours)
1. Create `internal/artwork/musicbrainz.go`
2. Implement Cover Art Archive API client
3. Add MBID lookup for albums/artists
4. Implement fallback logic

### Phase 5: Caching System (2 hours)
1. Create `internal/artwork/cache.go`
2. Implement file-based cache in `~/.cache/navitone-cli/artwork/`
3. Add cache expiration and cleanup
4. Optimize for performance

### Phase 6: Testing & Refinement (4 hours)
1. Test with various terminal sizes
2. Validate artwork quality and performance
3. Error handling and graceful fallbacks
4. Documentation updates

## File Structure

```
internal/
├── artwork/
│   ├── converter.go      # ASCII conversion logic
│   ├── musicbrainz.go    # MusicBrainz API client
│   ├── cache.go          # Artwork caching system
│   └── manager.go        # Artwork management coordination
├── config/
│   └── config.go         # Updated with artwork config
├── models/
│   ├── app.go            # Updated with artwork fields
│   └── config_form.go    # Updated with artwork form field
└── views/
    └── main.go           # Updated with artwork rendering
```

## Technical Specifications

### ASCII Art Conversion Settings
- **Default Size**: 40x20 characters
- **Color Support**: 8-bit terminal colors
- **Character Set**: Standard ASCII with density mapping
- **Fallback**: Monochrome for limited terminals

### Cache Strategy
- **Location**: `~/.cache/navitone-cli/artwork/`
- **Format**: `{album_id}_{size}.txt` for ASCII art files
- **Expiration**: 30 days or config change
- **Size Limit**: 10MB total cache size

### API Integration
- **Rate Limiting**: 1 request per second to MusicBrainz
- **Timeout**: 5 seconds for image downloads
- **Retry Logic**: 3 attempts with exponential backoff
- **Error Handling**: Graceful degradation if artwork unavailable

## Configuration Options

### Config File (`config.toml`)
```toml
[ui]
show_album_art = true
artwork_size = "medium"     # small, medium, large
artwork_color = true        # enable color ASCII art
artwork_cache_days = 30     # cache expiration
```

### UI Controls
- **Config Tab**: Toggle "Show Artwork" checkbox
- **Runtime**: Artwork respects current config state
- **Performance**: Artwork generation happens asynchronously

## Layout Specifications

### Albums Tab with Artwork
```
💿 Albums

[2023] Artist Name - Album Title             12 tracks (45 plays) ← Selected
[2022] Another Artist - Different Album      8 tracks (23 plays)
[2021] Third Artist - Another Album          15 tracks (67 plays)

    ████████████████████████████████████████
    ██████████████████░░░░░░██████████████░░
    ████████████░░░░░░░░░░░░░░░░░░████████░░
    ██████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░██░░
    ████░░░░░░░░░░████████░░░░░░░░░░░░░░░░░░
    ████░░░░░░████████████████░░░░░░░░░░░░░░    Album: Album Title
    ████░░░░██████████████████░░░░░░░░░░░░░░    Artist: Artist Name  
    ████░░░░██████████████████░░░░░░░░░░░░░░    Year: 2023
    ████░░░░████████████████░░░░░░░░░░░░░░░░    Tracks: 12
    ██████░░░░░░████████░░░░░░░░░░░░░░░░██░░
    ████████████░░░░░░░░░░░░░░░░░░████████░░
    ██████████████████░░░░░░██████████████░░
    ████████████████████████████████████████

Showing 1-15 of 245 albums                    Press M for more
```

### Space Management
- **Normal View**: 25 album entries visible
- **With Artwork**: 15 album entries + ASCII art (14 lines)
- **Minimum Height**: 30 lines total for artwork display
- **Responsive**: Artwork hidden if insufficient space

## Error Handling

### Graceful Degradation
1. **Config Disabled**: No artwork processing
2. **Network Issues**: Use cached artwork if available
3. **Invalid Images**: Skip artwork, show text only
4. **Small Terminal**: Hide artwork, maintain full functionality
5. **Conversion Errors**: Log and continue without artwork

### User Feedback
- **Loading State**: "Loading artwork..." indicator
- **Network Errors**: Brief status message in log area
- **Cache Status**: Transparent caching with optional debug info

## Performance Considerations

### Optimization Strategies
1. **Lazy Loading**: Generate artwork only for visible/selected items
2. **Background Processing**: Non-blocking artwork generation
3. **Smart Caching**: Aggressive caching with size limits
4. **Quality Scaling**: Adjust ASCII art quality based on terminal size

### Resource Management
- **Memory**: Limit concurrent conversions to 3
- **Network**: Rate limit API calls to be respectful
- **Storage**: Automatic cache cleanup and rotation
- **CPU**: Use efficient ASCII conversion algorithms

## Testing Strategy

### Manual Testing
1. **Various Terminal Sizes**: 80x24, 120x30, 200x50
2. **Network Conditions**: Online, offline, slow connection
3. **Different Albums**: With/without cover art, various art styles
4. **Config Changes**: Toggle on/off, different sizes

### Integration Points
- Config form validation and persistence
- Album/Artist tab navigation with artwork
- Modal interactions (ensure artwork doesn't interfere)
- Performance with large music libraries

## Future Enhancements

### Phase 2 Possibilities
1. **Braille Art**: Higher resolution using Unicode braille patterns
2. **Color Themes**: Artwork colors matching UI theme
3. **Animation**: Subtle transitions when switching selections
4. **User Uploads**: Allow custom artwork additions
5. **Artist Images**: Extend to artist photos in Artists tab

### Advanced Features
- **Playlist Artwork**: Composite or dominant album art for playlists
- **Now Playing**: Large artwork in dedicated view mode
- **Export Options**: Save ASCII art to files
- **Custom Characters**: User-defined character sets for conversion

## Implementation Checklist

### Phase 1: Config System
- [ ] Add ShowArtworkField to ConfigFormField enum
- [ ] Update IsCheckboxField() and GetCheckboxValue() functions  
- [ ] Update SetFieldValue() and ToggleCheckbox() functions
- [ ] Test config toggle functionality

### Phase 2: ASCII Conversion
- [ ] Create internal/artwork/ package structure
- [ ] Add ascii-image-converter dependency to go.mod
- [ ] Implement converter.go with basic functionality
- [ ] Add configuration options for size/quality

### Phase 3: UI Integration  
- [ ] Modify renderAlbumsTab() for artwork display
- [ ] Modify renderArtistsTab() for artwork display
- [ ] Add artwork area calculation logic
- [ ] Implement smart viewport adjustment

### Phase 4: MusicBrainz API
- [ ] Create musicbrainz.go API client
- [ ] Implement MBID lookup functionality
- [ ] Add Cover Art Archive integration
- [ ] Implement fallback logic

### Phase 5: Caching
- [ ] Create cache.go with file-based caching
- [ ] Add cache directory management
- [ ] Implement expiration and cleanup logic
- [ ] Add cache configuration options

### Phase 6: Testing
- [ ] Test various terminal sizes and conditions
- [ ] Validate artwork quality and performance
- [ ] Comprehensive error handling
- [ ] Update documentation and README

## Success Criteria

1. **Functionality**: Config toggle controls artwork display correctly
2. **Performance**: No noticeable lag in tab navigation 
3. **Reliability**: Graceful handling of network issues and missing artwork
4. **Usability**: Artwork enhances experience without disrupting workflow
5. **Compatibility**: Works across different terminal sizes and capabilities

## Completion Timeline: 2-3 Days

This implementation will significantly enhance the visual appeal and user experience of Navitone-CLI while maintaining its core terminal-based philosophy and performance characteristics.