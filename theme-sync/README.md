# Navitone Omarchy Theme Integration

Automatically synchronize [Navitone-CLI](https://github.com/yourusername/navitone-cli) colors with your current [Omarchy](https://github.com/aredl/omarchy) theme in real-time.

> **Note**: This is a personal integration for individual use and is **not part of the main Navitone project**. It's gitignored and separate from the core application.

## Features

- ✅ **Real-time theme synchronization** - Colors update automatically when you switch Omarchy themes
- ✅ **Universal theme support** - Works with JSON and TOML theme formats
- ✅ **Smart color mapping** - Maps Omarchy colors to Navitone's UI elements
- ✅ **Desktop notifications** - Get notified when colors update
- ✅ **Systemd integration** - Runs as a background service
- ✅ **Manual sync option** - Force update when needed

## Quick Setup

```bash
# Navigate to the theme-sync directory
cd theme-sync/

# Run the automated setup
./setup.sh
```

The setup script will:
1. Check for Omarchy installation and current theme
2. Install missing dependencies (inotify-tools, python-toml)
3. Create and enable a systemd user service
4. Test the initial theme sync
5. Start monitoring for theme changes

## Manual Usage

### One-time Theme Sync
```bash
# Extract current Omarchy theme and update Navitone config
./extract-omarchy-colors.py
```

### Start Manual Monitoring
```bash
# Run in foreground (useful for debugging)
./monitor-theme.sh
```

## How It Works

1. **Monitors** `~/.config/omarchy/current/theme` for changes using inotifywait
2. **Extracts** colors from theme files (`custom_theme.json` or `alacritty.toml`)
3. **Maps** Omarchy colors to Navitone UI elements:
   - `blue/cyan` → Accent colors (headers, highlights)
   - `magenta` → Secondary colors (selections)
   - `green` → Success indicators
   - `yellow` → Warning colors
   - `red` → Error colors
4. **Updates** Navitone's `~/.config/navitone-cli/config.toml`
5. **Notifies** you to restart Navitone to see changes

## Color Mapping

| Omarchy Color | Navitone Usage |
|---------------|----------------|
| `blue`        | Primary accent (headers, active elements) |
| `cyan`        | Secondary accent (highlights) |
| `magenta`     | Selection indicators |
| `green`       | Success states, play indicators |
| `yellow`      | Warning states |
| `red`         | Error states, stop indicators |
| `background`  | Terminal background (preserved) |
| `foreground`  | Text color |

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

## Supported Themes

- ✅ **JSON themes** with `custom_theme.json` (complex themes)
- ✅ **TOML themes** with `alacritty.toml` (standard themes)
- ✅ **Multiple hex formats** (`#xxxxxx`, `0xxxxxxx`)
- ✅ **Section-aware parsing** (`[colors.normal]`, `[colors.bright]`)
- ✅ **All standard Omarchy themes**

## Usage Workflow

1. Change your Omarchy theme (e.g., `omarchy dracula`)
2. Receive desktop notification: *"Navitone Theme Updated - Restart Navitone to see changes"*
3. Restart Navitone to see the new colors
4. Enjoy synchronized theming!

## Files

- `extract-omarchy-colors.py` - Core color extraction and config updating
- `monitor-theme.sh` - Background monitoring script
- `setup.sh` - Automated installation and service setup
- `theme-monitor.log` - Activity log file

## Requirements

- [Omarchy](https://github.com/aredl/omarchy) theme manager
- `inotify-tools` package (auto-installed by setup)
- Python 3.6+ with `toml` module (auto-installed by setup)

## Troubleshooting

### Service not starting
```bash
# Check service logs
journalctl --user -u navitone-theme-monitor -f
```

### Theme not updating
```bash
# Test manual sync
./extract-omarchy-colors.py

# Check if Omarchy theme is selected
ls -la ~/.config/omarchy/current/theme
```

### Missing dependencies
```bash
# Reinstall dependencies
sudo apt install inotify-tools  # Ubuntu/Debian
pip3 install --user toml
```

## Uninstall

```bash
# Stop and disable service
systemctl --user stop navitone-theme-monitor
systemctl --user disable navitone-theme-monitor

# Remove service file
rm ~/.config/systemd/user/navitone-theme-monitor.service

# Remove theme-sync directory
rm -rf theme-sync/
```

## License

Same as main Navitone project (MIT License)