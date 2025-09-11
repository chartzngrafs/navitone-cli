# Scrobbling Migration Plan: Server-Side First Approach

## Overview
Migrate from client-side scrobbling configuration to Navidrome's native server-side scrobbling, similar to how Feishin and other clients handle it. This eliminates the need for users to configure Last.fm API keys and ListenBrainz tokens directly in the client.

## Current Implementation Problems
- Users must obtain their own Last.fm API keys/secrets and ListenBrainz tokens
- Complex multi-step authentication process
- Duplicates functionality that Navidrome already provides
- Inconsistent with other Navidrome clients (Feishin, web UI)
- Increases barrier to entry for new users

## Navidrome's Native Scrobbling Support
- **API Endpoint**: `/rest/scrobble` (already implemented in client)
- **User Field**: `ScrobblingEnabled` in user profile
- **Server Config**: Admin configures API keys server-wide
- **User Setup**: Users link accounts through Navidrome web UI

## Migration Strategy

### Phase 1: Server Capability Detection
**Goal**: Detect and display server-side scrobbling availability

**Implementation**:
1. Add method to check user's `ScrobblingEnabled` field during auth
2. Query server capabilities for scrobbling configuration status
3. Update Config tab UI to show server scrobbling status
4. Display guidance message for users based on server capabilities

**Files to Modify**:
- `internal/api/navidrome/client.go` - Add capability detection
- `internal/models/config_form.go` - Update UI to show server status
- `internal/controllers/config.go` - Add server status checks

### Phase 2: Hybrid Scrobbling System
**Goal**: Support both server-side and client-side scrobbling with user choice

**Configuration Structure**:
```toml
[scrobbling]
method = "auto"  # "auto", "server", "client", "disabled"

# Existing client-side configs (maintained for fallback/advanced users)
[scrobbling.lastfm]
enabled = false
username = ""
password = ""
api_key = ""
secret = ""

[scrobbling.listenbrainz]
enabled = false
token = ""
```

**Scrobbling Methods**:
- **Auto**: Use server-side if available and enabled, fallback to client-side
- **Server**: Use Navidrome's native scrobbling only
- **Client**: Use current direct API approach (for advanced users)
- **Disabled**: No scrobbling

**Implementation**:
1. Add method selection to scrobbling manager
2. Implement server-side scrobbling route using existing `Client.Scrobble()`
3. Update routing logic in audio managers
4. Add fallback handling for server unavailability

**Files to Modify**:
- `pkg/scrobbling/manager.go` - Add method routing logic
- `pkg/scrobbling/types.go` - Add method enum and server capabilities
- `internal/config/config.go` - Update config structure
- `internal/audio/mpv/manager.go` - Update scrobbling calls
- `internal/models/config_form.go` - Add method selection UI

### Phase 3: Enhanced User Experience
**Goal**: Streamline setup process for most users

**Features**:
1. **Smart Detection**: Auto-configure based on server capabilities
2. **Setup Wizard**: Guide users through optimal scrobbling setup
3. **Status Indicators**: Clear feedback on scrobbling state
4. **Connection Testing**: Test both server and client-side scrobbling

**UI Enhancements**:
- Show "✅ Server scrobbling available" or "❌ Server scrobbling not configured"
- Hide complex API key fields when server-side is available and preferred
- Add "Test Scrobbling" button for both methods
- Provide setup instructions based on chosen method

## Implementation Details

### 1. Server Capability Detection
```go
// Add to navidrome client
type ScrobblingCapabilities struct {
    ServerScrobblingAvailable bool
    UserScrobblingEnabled     bool
    SupportedServices        []string
}

func (c *Client) GetScrobblingCapabilities(ctx context.Context) (*ScrobblingCapabilities, error) {
    // Check user profile for ScrobblingEnabled field
    // Query server configuration if possible
}
```

### 2. Enhanced Scrobbling Manager
```go
type ScrobblingMethod string

const (
    MethodAuto     ScrobblingMethod = "auto"
    MethodServer   ScrobblingMethod = "server"
    MethodClient   ScrobblingMethod = "client"
    MethodDisabled ScrobblingMethod = "disabled"
)

type Manager struct {
    method           ScrobblingMethod
    navidromeClient  *navidrome.Client
    // ... existing fields
}

func (m *Manager) ScrobbleTrack(track ScrobbleTrack, submission bool) error {
    switch m.method {
    case MethodServer:
        return m.scrobbleViaServer(track, submission)
    case MethodClient:
        return m.scrobbleViaClient(track, submission)
    case MethodAuto:
        return m.scrobbleAuto(track, submission)
    default:
        return nil // disabled
    }
}
```

### 3. Configuration UI Updates
```go
// Add to config form model
type ScrobblingMethodOption struct {
    Value       ScrobblingMethod
    Label       string
    Description string
    Available   bool
}

// Update form to show server capabilities and guide user choice
```

## Benefits

### For Users
- **Simplified Setup**: No need to obtain API keys for most users
- **Consistent Experience**: Matches behavior of other Navidrome clients
- **Reduced Configuration**: Single checkbox vs multiple API fields
- **Better Reliability**: Server handles authentication and retry logic

### For Administrators
- **Centralized Control**: Configure scrobbling once for all clients
- **User Management**: Control which users can scrobble
- **API Key Management**: Single set of API keys for all users
- **Consistent Configuration**: Same setup across all client applications

### For Advanced Users
- **Flexibility Maintained**: Can still use direct API configuration
- **Multiple Services**: Support services not configured server-side
- **Custom Settings**: Fine-tune client-side behavior if needed

## Migration Path for Existing Users

### Automatic Migration
1. **Detect existing client-side config**: Check if users have API keys configured
2. **Check server capabilities**: Query server for scrobbling support
3. **Suggest optimal setup**: Recommend server-side if available
4. **Preserve existing setup**: Don't break current configurations

### Migration Messages
- "Server scrobbling is now available! Switch to simplified setup?"
- "Keep current configuration or migrate to server-side scrobbling?"
- Clear explanation of benefits of each approach

## Testing Strategy

### Test Scenarios
1. **Server scrobbling available + enabled**: Should use server-side
2. **Server scrobbling available + disabled**: Should respect user choice
3. **Server scrobbling unavailable**: Should fallback to client-side
4. **Network issues**: Should handle gracefully with retry logic
5. **Mixed configuration**: Test auto mode with various server states

### Validation
- Test with actual Navidrome server with scrobbling configured
- Verify Last.fm and ListenBrainz integration through server
- Ensure existing client-side configurations continue working
- Test migration scenarios and fallback behavior

## Timeline

### Phase 1 (1-2 days)
- Implement server capability detection
- Update Config tab to show server status
- Basic UI improvements

### Phase 2 (2-3 days)
- Implement hybrid scrobbling system
- Add method selection and routing logic
- Update configuration structure

### Phase 3 (1-2 days)
- Enhanced UI/UX improvements
- Setup wizard and guidance
- Testing and polish

**Total Estimated Time**: 4-7 days

## Success Criteria
1. New users can enable scrobbling with single checkbox (if server configured)
2. Existing users' configurations continue working without changes
3. Advanced users can still use client-side APIs if preferred
4. Clear feedback on scrobbling status and setup options
5. Seamless fallback behavior when server scrobbling unavailable

## Future Considerations
- **Remove client-side entirely**: Once server-side adoption is high
- **Additional services**: Support services beyond Last.fm/ListenBrainz if server adds them
- **Real-time feedback**: Show scrobbling success/failure in UI
- **Statistics**: Track scrobbling success rates and method usage