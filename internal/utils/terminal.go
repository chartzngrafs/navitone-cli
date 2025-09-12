package utils

import (
	"os"
	"os/exec"
	"runtime"
)

// LaunchCavaInTerminal launches Cava audio visualizer in a new terminal window
func LaunchCavaInTerminal() error {
	// First check if cava is available
	if err := checkCavaInstalled(); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "linux":
		return launchLinuxTerminal()
	case "darwin":
		return launchMacOSTerminal()
	case "windows":
		return launchWindowsTerminal()
	default:
		return launchFallbackTerminal()
	}
}

// checkCavaInstalled verifies that cava is available in the system PATH
func checkCavaInstalled() error {
	_, err := exec.LookPath("cava")
	return err
}

// launchLinuxTerminal attempts to launch cava in common Linux terminals
func launchLinuxTerminal() error {
	terminals := [][]string{
		{"gnome-terminal", "--", "cava"},
		{"konsole", "-e", "cava"},
		{"xfce4-terminal", "-e", "cava"},
		{"mate-terminal", "-e", "cava"},
		{"lxterminal", "-e", "cava"},
		{"alacritty", "-e", "cava"},
		{"kitty", "cava"},
		{"st", "-e", "cava"},
		{"urxvt", "-e", "cava"},
		{"xterm", "-e", "cava"},
	}

	for _, termCmd := range terminals {
		if _, err := exec.LookPath(termCmd[0]); err == nil {
			cmd := exec.Command(termCmd[0], termCmd[1:]...)
			return cmd.Start()
		}
	}
	
	return launchFallbackTerminal()
}

// launchMacOSTerminal attempts to launch cava in macOS terminals
func launchMacOSTerminal() error {
	// Try iTerm2 first (more feature-rich)
	if _, err := exec.LookPath("osascript"); err == nil {
		// iTerm2 AppleScript
		iterm2Script := `tell application "iTerm"
			create window with default profile
			tell current session of current window
				write text "cava"
			end tell
		end tell`
		
		cmd := exec.Command("osascript", "-e", iterm2Script)
		if err := cmd.Start(); err == nil {
			return nil
		}
		
		// Fall back to Terminal.app
		terminalScript := `tell application "Terminal"
			activate
			do script "cava"
		end tell`
		
		cmd = exec.Command("osascript", "-e", terminalScript)
		return cmd.Start()
	}
	
	return launchFallbackTerminal()
}

// launchWindowsTerminal attempts to launch cava in Windows terminals
func launchWindowsTerminal() error {
	terminals := [][]string{
		{"wt", "cava"}, // Windows Terminal
		{"cmd", "/c", "start", "cmd", "/k", "cava"},
		{"powershell", "-Command", "Start-Process", "powershell", "-ArgumentList", "'-NoExit', '-Command', 'cava'"},
	}

	for _, termCmd := range terminals {
		if _, err := exec.LookPath(termCmd[0]); err == nil {
			cmd := exec.Command(termCmd[0], termCmd[1:]...)
			return cmd.Start()
		}
	}
	
	return launchFallbackTerminal()
}

// launchFallbackTerminal tries a simple approach as last resort
func launchFallbackTerminal() error {
	// Try environment variable first
	if term := os.Getenv("TERMINAL"); term != "" {
		cmd := exec.Command(term, "-e", "cava")
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	
	// System-specific fallbacks
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("x-terminal-emulator", "-e", "cava")
		return cmd.Start()
	case "darwin":
		cmd := exec.Command("open", "-a", "Terminal", "--args", "cava")
		return cmd.Start()
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "cava")
		return cmd.Start()
	}
	
	// Last resort - try running cava directly (won't work well but at least tries)
	cmd := exec.Command("cava")
	return cmd.Start()
}