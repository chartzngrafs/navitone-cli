#!/bin/bash

# Navitone Omarchy Theme Integration Setup
# Sets up automatic theme synchronization between Omarchy and Navitone

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="navitone-theme-monitor"
SERVICE_FILE="$HOME/.config/systemd/user/${SERVICE_NAME}.service"

echo "🎨 Navitone Omarchy Theme Integration Setup"
echo "==========================================="

# Check if Omarchy is installed
check_omarchy() {
    echo "Checking Omarchy installation..."
    if [ ! -d "$HOME/.config/omarchy" ]; then
        echo "❌ Omarchy not found at ~/.config/omarchy"
        echo "Please install Omarchy first: https://github.com/aredl/omarchy"
        exit 1
    fi

    if [ ! -L "$HOME/.config/omarchy/current/theme" ]; then
        echo "⚠️  No current theme selected in Omarchy"
        echo "Please select a theme in Omarchy first, then run this setup again"
        exit 1
    fi

    current_theme=$(basename "$(readlink "$HOME/.config/omarchy/current/theme")")
    echo "✅ Omarchy installed with current theme: $current_theme"
}

# Check dependencies
check_dependencies() {
    echo "Checking dependencies..."

    # Check inotify-tools
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

    # Check Python and toml module
    if ! command -v python3 &> /dev/null; then
        echo "❌ python3 not found"
        exit 1
    fi

    # Note: Script now uses built-in TOML parser, no external dependencies needed
    echo "Python toml module not required (using built-in parser)"

    echo "✅ All dependencies satisfied"
}

# Create systemd service
create_service() {
    echo "Creating systemd user service..."

    # Create systemd user directory if it doesn't exist
    mkdir -p "$(dirname "$SERVICE_FILE")"

    # Create service file
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Navitone Omarchy Theme Monitor
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
        echo "This is normal if you haven't run Navitone yet to create the config"
    fi
}

# Show status and instructions
show_status() {
    echo ""
    echo "🎉 Setup Complete!"
    echo "=================="
    echo ""
    echo "The theme monitor service is now running and will:"
    echo "• Automatically detect when you change Omarchy themes"
    echo "• Update Navitone's color configuration"
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
    echo "Note: Restart Navitone after theme changes to see new colors"
}

# Main setup function
main() {
    check_omarchy
    check_dependencies
    create_service
    enable_service
    test_sync
    show_status
}

# Handle interruption
cleanup() {
    echo ""
    echo "Setup interrupted"
    exit 1
}

trap cleanup SIGINT SIGTERM

# Run setup
main "$@"