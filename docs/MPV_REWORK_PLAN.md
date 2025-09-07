# Navitone MPV Backend Rework Plan

## Current Issues with Direct Audio Playback

### Fundamental Problems
1. **Seeking Limitations**: HTTP Range seeking fails for compressed formats (FLAC, MP3, OGG)
2. **Decoder Frame Boundaries**: Can't start playback from arbitrary byte positions in compressed audio
3. **Limited Format Support**: Manual decoder implementation is complex and error-prone
4. **Position Tracking Issues**: Time-based position estimation is inaccurate
5. **Audio Buffer Management**: Manual audio buffer handling causes complexity

### Why Current Approach Fails
- **Compressed Audio Streams**: Need to start from frame boundaries, not arbitrary bytes
- **Network Streaming**: HTTP Range requests work for bytes, not audio frames
- **Decoder Complexity**: Each format requires different decoding logic
- **Real-time Constraints**: Audio playback has strict timing requirements

## Proposed Solution: MPV Backend

### Why MPV?
MPV is a mature, battle-tested media player with:
- **Native HTTP Streaming**: Handles network streams expertly
- **Universal Format Support**: Plays virtually any audio format
- **Proper Seeking**: Frame-accurate seeking in all formats
- **Buffer Management**: Intelligent buffering and network handling
- **JSON IPC Interface**: Perfect for programmatic control
- **Cross-platform**: Works on Linux, macOS, Windows

### Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Navitone UI   │    │  MPV Process    │    │  Navidrome      │
│   (BubbleTea)   │    │                 │    │   Server        │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ JSON IPC Commands    │ HTTP Audio Streams   │
          ├─────────────────────►│◄─────────────────────┤
          │                      │                      │
          │ Status/Events        │                      │
          │◄─────────────────────┤                      │
          │                      │                      │
```

## Implementation Plan

### Phase 1: MPV Integration Layer

#### 1.1 MPV Process Management
```go
// internal/audio/mpv/process.go
type MPVProcess struct {
    process    *exec.Cmd
    socketPath string
    logPath    string
    ipc        *IPCClient
}

func NewMPVProcess(socketPath string) *MPVProcess
func (m *MPVProcess) Start(args []string) error
func (m *MPVProcess) Stop() error
func (m *MPVProcess) IsRunning() bool
```

#### 1.2 IPC Communication Layer
```go
// internal/audio/mpv/ipc.go
type IPCClient struct {
    conn     net.Conn
    requests map[int]chan IPCResponse
    events   chan MPVEvent
}

type MPVCommand struct {
    Command   []interface{} `json:"command"`
    RequestID int           `json:"request_id,omitempty"`
}

type MPVEvent struct {
    Event string      `json:"event"`
    Data  interface{} `json:",inline"`
}
```

#### 1.3 Audio Manager Replacement
```go
// internal/audio/mpv/manager.go
type MPVManager struct {
    process    *MPVProcess
    ipc        *IPCClient
    
    // State management
    queue      []models.Track
    currentIdx int
    isPlaying  bool
    position   time.Duration
    duration   time.Duration
    
    // Callbacks
    stateCallback func(*models.AppState)
    logCallback   func(string)
}
```

### Phase 2: Core Functionality

#### 2.1 Playback Control
```go
// Play a track
func (m *MPVManager) PlayTrack(track models.Track) error {
    streamURL := m.navidromeClient.GetStreamURL(track.ID)
    return m.ipc.SendCommand("loadfile", streamURL)
}

// Queue management
func (m *MPVManager) AddToQueue(track models.Track) error {
    streamURL := m.navidromeClient.GetStreamURL(track.ID)
    return m.ipc.SendCommand("loadfile", streamURL, "append-play")
}

// Seeking (finally works!)
func (m *MPVManager) SeekForward(seconds int) error {
    return m.ipc.SendCommand("seek", seconds, "relative")
}

func (m *MPVManager) SeekBackward(seconds int) error {
    return m.ipc.SendCommand("seek", -seconds, "relative")
}

func (m *MPVManager) SeekToPosition(position time.Duration) error {
    return m.ipc.SendCommand("seek", position.Seconds(), "absolute")
}
```

#### 2.2 State Synchronization
```go
// Event handling from MPV
func (m *MPVManager) handleMPVEvent(event MPVEvent) {
    switch event.Event {
    case "file-loaded":
        m.updateTrackInfo()
    case "playback-time":
        m.position = time.Duration(event.Data.(float64)) * time.Second
        m.notifyPositionUpdate()
    case "end-file":
        m.handleTrackEnd()
    case "pause":
        m.isPlaying = false
        m.notifyStateChange()
    }
}
```

### Phase 3: Advanced Features

#### 3.1 Enhanced Queue Management
- **Gapless Playback**: MPV handles seamless track transitions
- **Crossfading**: Optional crossfade between tracks
- **Smart Buffering**: MPV pre-loads next tracks automatically

#### 3.2 Audio Processing
```go
// Audio effects and processing
func (m *MPVManager) SetReplayGain(mode string) error {
    return m.ipc.SendCommand("set_property", "replaygain", mode)
}

func (m *MPVManager) SetEqualizer(bands []float64) error {
    // Configure MPV's audio filters
}
```

#### 3.3 Advanced Seeking
```go
// Frame-accurate seeking
func (m *MPVManager) SeekToFrame(frame int) error {
    return m.ipc.SendCommand("seek", frame, "absolute", "keyframes")
}

// Chapter/marker support
func (m *MPVManager) SeekToChapter(chapter int) error {
    return m.ipc.SendCommand("set_property", "chapter", chapter)
}
```

## Implementation Benefits

### Immediate Fixes
✅ **Perfect Seeking**: Frame-accurate seeking in all formats  
✅ **Format Support**: Universal audio format compatibility  
✅ **Network Handling**: Robust HTTP streaming with retry logic  
✅ **Buffer Management**: Intelligent buffering prevents dropouts  
✅ **Gapless Playback**: Seamless track transitions  

### Enhanced Features
✅ **Audio Processing**: Built-in EQ, replay gain, effects  
✅ **Subtitle Support**: For audio with embedded text  
✅ **Playlist Management**: Native playlist handling  
✅ **Stream Information**: Rich metadata extraction  
✅ **Performance**: Optimized audio pipeline  

## File Structure Changes

```
internal/audio/
├── mpv/
│   ├── process.go      # MPV process management
│   ├── ipc.go         # JSON IPC communication
│   ├── manager.go     # High-level audio manager
│   ├── events.go      # Event handling and parsing
│   └── commands.go    # MPV command wrappers
├── legacy/            # Keep old implementation for reference
│   ├── player.go
│   ├── decoder.go
│   └── manager.go
└── manager.go         # Interface definition
```

## Migration Strategy

### Phase 1: Parallel Implementation
- Keep existing audio system running
- Implement MPV backend alongside
- Add feature flag to switch between backends

### Phase 2: Testing and Validation
- Comprehensive testing of MPV integration
- Performance comparison
- User acceptance testing

### Phase 3: Cutover
- Default to MPV backend
- Remove old audio implementation
- Clean up dependencies

## Dependencies

### New Dependencies
```go
// go.mod additions
// No new Go dependencies needed - uses standard library
// MPV communication via JSON over Unix socket
```

### System Requirements
- **MPV binary**: Must be available in system PATH
- **Unix Socket Support**: For IPC communication
- **Temp Directory**: For socket files

## Configuration

### MPV Configuration
```go
// internal/audio/mpv/config.go
type MPVConfig struct {
    SocketPath    string
    LogFile       string
    AudioDevice   string
    BufferSize    time.Duration
    EnableGapless bool
    ReplayGain    string
}
```

### Runtime Arguments
```bash
mpv --no-video \
    --idle \
    --input-ipc-server=/tmp/navitone_mpv \
    --log-file=/tmp/navitone_mpv.log \
    --audio-buffer=0.5 \
    --gapless-audio=yes \
    --replaygain=track
```

## Error Handling

### Robust Error Recovery
```go
// MPV process monitoring
func (m *MPVManager) monitorProcess() {
    for {
        if !m.process.IsRunning() {
            log.Warn("MPV process died, restarting...")
            m.restartMPV()
        }
        time.Sleep(5 * time.Second)
    }
}

// IPC connection recovery
func (m *MPVManager) ensureConnection() error {
    if !m.ipc.IsConnected() {
        return m.ipc.Reconnect()
    }
    return nil
}
```

## Testing Strategy

### Unit Tests
- MPV process lifecycle
- IPC command serialization
- Event parsing and handling

### Integration Tests  
- Full playback scenarios
- Seeking accuracy tests
- Queue management validation

### Performance Tests
- Memory usage comparison
- CPU usage benchmarks
- Audio latency measurements

## Timeline Estimate

### Phase 1 (Week 1-2): Foundation
- MPV process management
- Basic IPC communication
- Simple play/pause/stop

### Phase 2 (Week 2-3): Core Features
- Queue management
- Seeking implementation
- State synchronization

### Phase 3 (Week 3-4): Integration
- UI integration
- Error handling
- Testing and debugging

### Phase 4 (Week 4): Polish
- Performance optimization
- Documentation
- Migration cleanup

## Risk Mitigation

### Potential Issues
1. **MPV Dependency**: Requires MPV installation
2. **Process Management**: Handling crashes and restarts
3. **IPC Complexity**: JSON parsing and event handling

### Mitigation Strategies
1. **Fallback Mode**: Keep basic playback as fallback
2. **Health Monitoring**: Automatic process restart
3. **Comprehensive Testing**: Edge case validation

## Conclusion

Switching to MPV as the audio backend will solve all current playback issues and provide a robust, professional-grade audio system. The investment in rework will pay dividends in reliability, features, and user experience.

The naviterm project proves this approach works well for terminal-based music players. We can leverage their learnings while building a more integrated solution for Navitone.

**Recommendation**: Proceed with MPV backend implementation as it's the only viable solution for professional-quality audio playback with proper seeking support.