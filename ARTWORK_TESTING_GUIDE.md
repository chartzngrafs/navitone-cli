# 🎨 ASCII Artwork Feature - Testing Guide

## ✅ Issues Fixed

1. **Config Toggle Missing** → ✅ **FIXED** - Added "UI Settings" section to Config tab
2. **Artwork Not Displaying** → ✅ **FIXED** - Fixed state initialization and loading logic  
3. **No Tab Switch Loading** → ✅ **FIXED** - Added artwork loading when entering Albums/Artists tabs
4. **Navidrome URL Construction** → ✅ **FIXED** - Properly construct authenticated getCoverArt URLs

## 🧪 Testing Steps

### Step 1: Enable Artwork Feature
1. Launch Navitone: `./bin/navitone`
2. Navigate to **Config** tab (rightmost tab using Tab key)
3. Look for **"UI Settings"** section 
4. Find **"Show Artwork"** checkbox
5. Use **Enter** to toggle it **ON** (should show `[X]`)
6. Press **F2** to save configuration
7. You should see "Configuration saved successfully!" message

### Step 2: Test Albums Tab
1. Navigate to **Albums** tab (use Tab key)
2. Wait for albums to load from your Navidrome server
3. Use **↑↓ arrow keys** to navigate between albums
4. **Expected Result**: ASCII artwork should appear below the album list for each selected album
5. **First Load**: May take a few seconds (downloading & converting)
6. **Subsequent**: Should be instant (cached)

### Step 3: Test Artists Tab  
1. Navigate to **Artists** tab
2. Use **↑↓ arrow keys** to navigate between artists
3. **Expected Result**: ASCII artwork should appear below the artist list
4. **Note**: Artist artwork uses MusicBrainz fallback (may be slower or not available for all artists)

### Step 4: Test Zen Mode (Shift+Z)
1. Press **Shift+Z** (or **Z**) anywhere in the application
2. **Expected Layout**:
   - **Centered Full-Color ASCII Album Art** (bypasses "Show Artwork" config toggle)
   - **Track & Playing Info**: `🎵 Track Title` & `👤 Artist • 💿 Album (Year)`
   - **Playing State & Progress**: `▶ Playing  •  Vol: 80%  •  🔀 Shuffle` + Progress timestamp (`01:23 / 03:45`) and progress bar (`[▓▓▓▓▓░░░░░]`)
   - **Hidden Chrome**: Tabs, sidebars, track lists, and log areas are hidden for a clean experience
3. **Playback Controls**: Media keys (`Space` to play/pause, `Alt+Right`/`Alt+Left` to skip, `Shift+Up`/`Shift+Down` for volume) continue working while in Zen Mode
4. Press **Shift+Z** again to exit Zen Mode and restore the standard UI layout

## 🔍 What You Should See

### Config Tab (with quality options):
```
Configuration

Navidrome Server Settings
→ Server URL: [your-server]
  Username: [your-username]  
  Password: ••••••••

UI Settings
→ Show Artwork: [X]      ← Enable artwork display
  Artwork Quality: high   ← NEW! Quality level
  Artwork Color: [ ]      ← NEW! Enable colors  
  Artwork Size: medium    ← NEW! Size setting

Audio Settings
→ Volume: 100%
  Audio Device: Auto-detect
  Buffer Size: 4096
```

### Albums Tab (with artwork enabled):
```
💿 Albums

[2023] Artist Name - Album Title             12 tracks (45 plays) ← Selected
[2022] Another Artist - Different Album      8 tracks (23 plays)
[2021] Third Artist - Another Album          15 tracks (67 plays)

🎨 Artist Name - Album Title (2023)

    ████████████████████████████████████████
    ██████░░░░░░░░░░░░░░░░░░░░░░░░░░████████
    ████░░░░░░░░░░██████░░░░░░░░░░░░░░░░████
    ██░░░░░░░░██████████████░░░░░░░░░░░░░░██
    ░░░░░░░░████████████████████░░░░░░░░░░░░
    ░░░░░░██████████████████████░░░░░░░░░░░░
    ░░░░░░██████████████████████░░░░░░░░░░░░
    ██░░░░░░░░██████████████░░░░░░░░░░░░░░██
    ████░░░░░░░░░░██████░░░░░░░░░░░░░░░░████
    ██████░░░░░░░░░░░░░░░░░░░░░░░░░░████████
    ████████████████████████████████████████

Showing 1-15 of 245 albums

╭──────────────────────────────────────────────────────────────────────────────╮
│ ▶ Playing  🎵 Song Title — Artist (Album Name)     Vol: 80% • 🔀 ON • Q: 12  │
│ 01:23  [ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░ ]  03:45                    │
│ ⚡ Loaded artwork for Album          Enter Tracks • A Queue • Shift+S Sort...│
╰──────────────────────────────────────────────────────────────────────────────╯
```

## 🐛 Troubleshooting

### Issue: Config toggle not visible
- **Check**: Config tab should show "UI Settings" section
- **If missing**: Rebuild with `go build -o bin/navitone ./cmd/navitone`

### Issue: No artwork appears
1. **Check**: Is "Show Artwork" checkbox enabled (`[X]`)?
2. **Check**: Log area at bottom - look for artwork loading messages
3. **Check**: Do your albums have cover art URLs from Navidrome?
4. **Check**: Network connectivity for MusicBrainz fallback

### Issue: Artwork loading is slow
- **Expected**: First load takes time (downloading + converting)
- **Expected**: Cached loads should be instant
- **Fallback**: MusicBrainz requests may be slower

### Issue: Some albums have no artwork
- **Expected**: Not all albums have cover art available
- **Fallback Chain**: Navidrome → MusicBrainz → None  
- **Check**: Log messages for specific errors

### Issue: "failed to convert artwork" errors
- **Likely**: Navidrome cover art URL construction issues → **FIXED**
- **Check**: Log should show full URLs like `https://your-server/rest/getCoverArt?...`
- **Previously**: Showed errors like "unable to open file: al-xyz..." → **RESOLVED**

## 📊 Expected Performance

- **Cache Location**: `~/.cache/navitone-cli/artwork/`
- **Cache Expiration**: 30 days
- **Image Size**: 40x20 characters (approx 15 lines)
- **First Load**: 2-10 seconds (depending on image size/network)
- **Cached Load**: Instant
- **UI Impact**: Reduces visible album/artist count from 25 to 15

## 🎨 Quality Optimization Guide

### Quality Level Comparison:
- **LOW**: 10 characters, basic detail, fastest performance
- **MEDIUM**: 69 characters, much better detail, good balance  
- **HIGH**: 69 chars + optimized mapping, excellent quality (recommended)
- **ULTRA**: Braille characters, maximum detail, requires UTF-8 terminal

### Size Recommendations:
- **Small**: 35x18 - Compact terminals, limited space
- **Medium**: 50x25 - Best balance of detail and space (recommended)
- **Large**: 70x35 - Large terminals, maximum detail

### Color vs Monochrome:
- **Monochrome**: Compatible with all terminals, clean appearance
- **Color**: 24-bit color support, much more realistic, requires modern terminal

### Performance Impact:
- Higher quality = larger file sizes but better visual result
- Ultra quality with braille = highest detail but slower conversion
- Color adds ANSI escape codes but modern terminals handle well

### Recommended Settings:
- **Best Balance**: High quality, Medium size, Monochrome
- **Maximum Quality**: Ultra quality, Large size, Color (if terminal supports)
- **Performance**: Medium quality, Small size, Monochrome

## ✅ Success Criteria

1. **Config Toggle**: ✅ "Show Artwork" appears in UI Settings
2. **Toggle Function**: ✅ Can enable/disable and save successfully  
3. **Albums Artwork**: ✅ ASCII art displays below album list on selection
4. **Artists Artwork**: ✅ ASCII art displays below artist list on selection
5. **Navigation**: ✅ Artwork changes when using ↑↓ keys
6. **Caching**: ✅ Second view of same album is instant
7. **Responsive**: ✅ UI adjusts item count to make room for artwork
8. **Error Handling**: ✅ App continues working if artwork fails

## 🎯 Test Results

Please test and report:

- [ ] Config toggle appears and works
- [ ] Albums tab shows artwork when navigating  
- [ ] Artists tab shows artwork when navigating
- [ ] Artwork is cached (second view is fast)
- [ ] UI layout adjusts properly
- [ ] No crashes or major performance issues

The ASCII artwork feature should now be fully functional! 🎵🎨