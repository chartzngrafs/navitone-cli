# Navitone Omarchy Quattro Theme Integration

Automatically synchronize [Navitone-CLI](https://github.com/yourusername/navitone-cli) colors with your current [Omarchy 4.0 Quattro](https://github.com/aredl/omarchy) desktop theme in real-time.

> **Note**: This integration supports Omarchy 4.0 Quattro 24-color palettes, Base24 schemes, and Cava vertical gradient palettes.

## Features

- ✅ **Real-time theme synchronization** - Colors update automatically when you switch Omarchy themes
- ✅ **Omarchy 4.0 Quattro support** - Multi-path resolution for `~/.local/state/omarchy/current/theme/` state directories
- ✅ **24-Color & Base24 Schemas** - Parses 24-color extended palettes (`color0`..`color23`) and Base24 schemes (`base00`..`base17`)
- ✅ **Vertical Cava Visualizer Palette** - Extracts `cava_theme` or generates background-filtered spectrum gradient colors
- ✅ **Live Hotkey Reload (`Shift+R`)** - Instant live theme refresh inside Navitone without restarting
- ✅ **Desktop notifications** - Get notified when colors update
- ✅ **Systemd integration** - Runs as a background monitoring service

## Quick Setup

```bash
# Navigate to the theme-sync directory
cd theme-sync/

# Run the automated setup
./setup.sh
```

The setup script will:
1. Check for Omarchy Quattro installation and current theme
2. Install missing dependencies (`inotify-tools`, `python-toml`)
3. Create and enable a systemd user service (`navitone-theme-monitor.service`)
4. Test the initial theme sync
5. Start monitoring for theme changes across `~/.local/state/omarchy/` and `~/.config/omarchy/`

## Manual Usage

### One-time Theme Sync
```bash
# Extract current Omarchy Quattro theme and update Navitone config
./extract-omarchy-colors.py
```

### Start Manual Monitoring
```bash
# Run in foreground (useful for debugging)
./monitor-theme.sh
```

## How It Works

1. **Monitors** `~/.local/state/omarchy/` and `~/.config/omarchy/` for filesystem events using `inotifywait`
2. **Extracts** colors from theme files (`colors.toml`, `cava_theme`, `custom_theme.json`, or `alacritty.toml`)
3. **Parses** 24-color Quattro palettes (`color0`..`color23`) and Base24 schemes (`base00`..`base17`)
4. **Maps** Omarchy colors to Navitone UI semantic roles:
   - `color4/color12/accent` → Accent & Primary colors
   - `color5/color13/secondary` → Selection indicators
   - `color2/color10/success` → Success & Play states
   - `color3/color11/warning` → Warning states
   - `color1/color9/error` → Error states
5. **Generates** `cava_palette` from `cava_theme` or filters background colors from `palette` for maximum contrast
6. **Updates** Navitone's `~/.config/navitone-cli/config.toml`
7. **Notifies** you to press `Shift+R` in Navitone to apply changes live!

## Color Mapping

| Omarchy Color Key | Navitone Usage |
|-------------------|----------------|
| `accent` / `color4` / `color12` | Primary accent (headers, active borders) |
| `secondary` / `color5` / `color13` | Selection indicators, highlights |
| `success` / `color2` / `color10` | Success states, play indicators |
| `warning` / `color3` / `color11` | Warning states, buffering |
| `error` / `color1` / `color9` | Error states, stop indicators |
| `color0` .. `color23` | Extended 24-color moving gradient palette |
| `gradient_color_1` .. `N` | Cava vertical spectrum visualizer gradient |
| `background` / `foreground` | Interface background & text colors |

## Service Management

The setup creates a systemd user service that runs automatically:

```bash
# Check service status
systemctl --user status navitone-theme-monitor

# Stop the service
systemctl --user stop navitone-theme-monitor

# Start the service
systemctl --user start navitone-theme-monitor

# Disable auto-start
systemctl --user disable navitone-theme-monitor

# View logs
tail -f theme-monitor.log
```

## Files

- `extract-omarchy-colors.py` - Core color extraction and config updating
- `monitor-theme.sh` - Background monitoring script
- `setup.sh` - Automated installation and service setup
- `theme-monitor.log` - Activity log file

## License

MIT License (Same as main Navitone project)