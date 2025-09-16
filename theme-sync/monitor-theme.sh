#!/bin/bash

# Navitone Omarchy Theme Monitor
# Watches for Omarchy theme changes and updates Navitone colors automatically

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
THEME_LINK="$HOME/.config/omarchy/current/theme"
LOG_FILE="$SCRIPT_DIR/theme-monitor.log"
COLOR_EXTRACTOR="$SCRIPT_DIR/extract-omarchy-colors.py"

# Check if required tools are installed
check_dependencies() {
    if ! command -v inotifywait &> /dev/null; then
        echo "Error: inotifywait not found. Install inotify-tools package."
        echo "Ubuntu/Debian: sudo apt install inotify-tools"
        echo "Arch: sudo pacman -S inotify-tools"
        exit 1
    fi

    if ! command -v python3 &> /dev/null; then
        echo "Error: python3 not found."
        exit 1
    fi

    # Note: Python script now uses built-in TOML parser, no external dependencies needed
    echo "Using built-in TOML parser (no external dependencies required)"

    if [ ! -f "$COLOR_EXTRACTOR" ]; then
        echo "Error: Color extractor script not found at $COLOR_EXTRACTOR"
        exit 1
    fi
}

# Send desktop notification
send_notification() {
    local title="$1"
    local message="$2"

    if command -v notify-send &> /dev/null; then
        notify-send -i "applications-multimedia" "$title" "$message"
    else
        echo "$title: $message"
    fi
}

# Update theme colors
update_theme() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] Theme change detected, updating Navitone colors..." | tee -a "$LOG_FILE"

    if python3 "$COLOR_EXTRACTOR"; then
        echo "[$timestamp] ✅ Navitone colors updated successfully" | tee -a "$LOG_FILE"
        send_notification "Navitone Theme Updated" "Colors synchronized with current Omarchy theme. Restart Navitone to see changes."
    else
        echo "[$timestamp] ❌ Failed to update Navitone colors" | tee -a "$LOG_FILE"
        send_notification "Navitone Theme Update Failed" "Could not extract colors from current theme"
    fi
}

# Main monitoring loop
main() {
    echo "Starting Navitone Omarchy theme monitor..."
    echo "Monitoring: $THEME_LINK"
    echo "Log file: $LOG_FILE"

    # Check dependencies
    check_dependencies

    # Initial theme sync
    if [ -L "$THEME_LINK" ]; then
        current_theme=$(basename "$(readlink "$THEME_LINK")")
        echo "Current theme: $current_theme"
        update_theme
    else
        echo "Warning: Omarchy theme symlink not found at $THEME_LINK"
        echo "Make sure Omarchy is installed and a theme is selected."
    fi

    # Start monitoring
    echo "Starting file system monitor..."
    while true; do
        # Wait for changes to the theme symlink
        inotifywait -e modify,move,create,delete -q "$HOME/.config/omarchy/current/" 2>/dev/null

        # Small delay to avoid multiple rapid events
        sleep 0.5

        # Check if theme link still exists and get new theme
        if [ -L "$THEME_LINK" ]; then
            new_theme=$(basename "$(readlink "$THEME_LINK")")
            if [ "$new_theme" != "$current_theme" ]; then
                current_theme="$new_theme"
                echo "Theme changed to: $current_theme"
                update_theme
            fi
        fi
    done
}

# Handle script termination
cleanup() {
    echo ""
    echo "Theme monitor stopped."
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Run main function
main "$@"