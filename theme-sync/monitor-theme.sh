#!/bin/bash

# Navitone Omarchy Quattro Theme Monitor
# Watches for Omarchy theme changes and updates Navitone colors automatically

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATE_DIR="$HOME/.local/state/omarchy"
CONFIG_DIR="$HOME/.config/omarchy"
LOG_FILE="$SCRIPT_DIR/theme-monitor.log"
COLOR_EXTRACTOR="$SCRIPT_DIR/extract-omarchy-colors.py"

check_dependencies() {
    if ! command -v inotifywait &> /dev/null; then
        echo "Error: inotifywait not found. Install inotify-tools package."
        exit 1
    fi

    if ! command -v python3 &> /dev/null; then
        echo "Error: python3 not found."
        exit 1
    fi

    if [ ! -f "$COLOR_EXTRACTOR" ]; then
        echo "Error: Color extractor script not found at $COLOR_EXTRACTOR"
        exit 1
    fi
}

send_notification() {
    local title="$1"
    local message="$2"

    if command -v notify-send &> /dev/null; then
        notify-send -i "applications-multimedia" "$title" "$message"
    else
        echo "$title: $message"
    fi
}

update_theme() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] Theme change detected, updating Navitone colors..." | tee -a "$LOG_FILE"

    if python3 "$COLOR_EXTRACTOR"; then
        echo "[$timestamp] ✅ Navitone colors updated successfully" | tee -a "$LOG_FILE"
        send_notification "Navitone Theme Updated" "Colors synchronized with current Omarchy Quattro theme. Press Shift+R in Navitone to reload live."
    else
        echo "[$timestamp] ❌ Failed to update Navitone colors" | tee -a "$LOG_FILE"
        send_notification "Navitone Theme Update Failed" "Could not extract colors from current theme"
    fi
}

get_theme_name() {
    for name_file in "$STATE_DIR/current/theme.name" "$STATE_DIR/theme.name" "$CONFIG_DIR/current/theme.name" "$CONFIG_DIR/theme.name" "$CONFIG_DIR/active_theme"; do
        if [ -f "$name_file" ]; then
            cat "$name_file"
            return
        fi
    done

    if [ -d "$STATE_DIR/current/theme" ]; then
        echo "firesky"
        return
    fi

    echo "unknown"
}

theme_exists() {
    if [ -d "$STATE_DIR" ] || [ -d "$CONFIG_DIR" ]; then
        return 0
    fi
    return 1
}

main() {
    echo "Starting Navitone Omarchy Quattro theme monitor..."
    echo "Monitoring: $STATE_DIR and $CONFIG_DIR"
    echo "Log file: $LOG_FILE"

    check_dependencies

    if theme_exists; then
        current_theme=$(get_theme_name)
        echo "Current theme: $current_theme"
        update_theme
    fi

    echo "Starting file system monitor..."
    while true; do
        inotifywait -r -e modify,move,create,delete -q "$STATE_DIR/" "$CONFIG_DIR/" 2>/dev/null

        sleep 0.5

        if theme_exists; then
            new_theme=$(get_theme_name)
            if [ "$new_theme" != "$current_theme" ]; then
                current_theme="$new_theme"
                echo "Theme changed to: $current_theme"
                update_theme
            fi
        fi
    done
}

cleanup() {
    echo ""
    echo "Theme monitor stopped."
    exit 0
}

trap cleanup SIGINT SIGTERM

main "$@"