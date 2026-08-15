#!/bin/bash

# Navitone Omarchy Quattro Theme Integration Setup
# Sets up automatic theme synchronization between Omarchy Quattro and Navitone

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="navitone-theme-monitor"
SERVICE_FILE="$HOME/.config/systemd/user/${SERVICE_NAME}.service"

echo "🎨 Navitone Omarchy Quattro Theme Integration Setup"
echo "=================================================="

# Check if Omarchy is installed
check_omarchy() {
    echo "Checking Omarchy installation..."
    if [ ! -d "$HOME/.config/omarchy" ]; then
        echo "❌ Omarchy not found at ~/.config/omarchy"
        echo "Please install Omarchy first: https://github.com/aredl/omarchy"
        exit 1
    fi

    local current_theme="unknown"
    for name_file in "$HOME/.config/omarchy/current/theme.name" "$HOME/.config/omarchy/theme.name" "$HOME/.config/omarchy/active_theme" "$HOME/.config/omarchy/current/theme/theme.name"; do
        if [ -f "$name_file" ]; then
            current_theme=$(cat "$name_file")
            break
        fi
    done

    if [ "$current_theme" = "unknown" ]; then
        local theme_dir="$HOME/.config/omarchy/current/theme"
        if [ -L "$theme_dir" ]; then
            current_theme=$(basename "$(readlink "$theme_dir")")
        elif [ -d "$theme_dir" ]; then
            current_theme="current"
        fi
    fi

    echo "✅ Omarchy Quattro installed with current theme: $current_theme"
}

# Check dependencies
check_dependencies() {
    echo "Checking dependencies..."

    if ! command -v inotifywait &> /dev/null; then
        echo "❌ inotifywait not found. Installing inotify-tools..."
        if command -v apt &> /dev/null; then
            sudo apt update && sudo apt install -y inotify-tools
        elif command -v pacman &> /dev/null; then
            sudo pacman -S inotify-tools
        elif command -v dnf &> /dev/null; then
            sudo dnf install inotify-tools
        else
            echo "Please install inotify-tools manually for your distribution"
            exit 1
        fi
    fi

    if ! command -v python3 &> /dev/null; then
        echo "❌ python3 not found"
        exit 1
    fi

    echo "✅ All dependencies satisfied"
}

# Create systemd service
create_service() {
    echo "Creating systemd user service..."

    mkdir -p "$(dirname "$SERVICE_FILE")"

    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Navitone Omarchy Quattro Theme Monitor
After=graphical-session.target

[Service]
Type=simple
ExecStart=$SCRIPT_DIR/monitor-theme.sh
Restart=always
RestartSec=5
Environment=DISPLAY=:0
Environment=XDG_RUNTIME_DIR=%i

[Install]
WantedBy=default.target
EOF

    echo "✅ Service file created at $SERVICE_FILE"
}

# Enable and start service
enable_service() {
    echo "Enabling and starting service..."
    systemctl --user daemon-reload
    systemctl --user enable "$SERVICE_NAME.service"
    systemctl --user start "$SERVICE_NAME.service"
    echo "✅ Service enabled and started"
}

# Test initial sync
test_sync() {
    echo "Testing initial theme sync..."
    if python3 "$SCRIPT_DIR/extract-omarchy-colors.py"; then
        echo "✅ Theme sync test successful"
    else
        echo "⚠️  Theme sync test failed - check if Navitone config exists"
    fi
}

# Show status and instructions
show_status() {
    echo ""
    echo "🎉 Setup Complete!"
    echo "=================="
    echo ""
    echo "The theme monitor service is now running and will:"
    echo "• Automatically detect when you change Omarchy Quattro themes"
    echo "• Update Navitone's 24-color configuration"
    echo "• Send desktop notifications when themes change"
    echo ""
    echo "Service Management:"
    echo "• Status:  systemctl --user status $SERVICE_NAME"
    echo "• Stop:    systemctl --user stop $SERVICE_NAME"
    echo "• Start:   systemctl --user start $SERVICE_NAME"
    echo "• Logs:    tail -f $SCRIPT_DIR/theme-monitor.log"
    echo ""
    echo "Manual sync: $SCRIPT_DIR/extract-omarchy-colors.py"
    echo ""
    echo "Note: Press Shift+R in Navitone to reload theme colors live on demand"
}

main() {
    check_omarchy
    check_dependencies
    create_service
    enable_service
    test_sync
    show_status
}

cleanup() {
    echo ""
    echo "Setup interrupted"
    exit 1
}

trap cleanup SIGINT SIGTERM

main "$@"