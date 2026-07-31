package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type hyprWindow struct {
	Address      string `json:"address"`
	Title        string `json:"title"`
	InitialTitle string `json:"initialTitle"`
	Class        string `json:"class"`
	InitialClass string `json:"initialClass"`
	Workspace    struct {
		ID int `json:"id"`
	} `json:"workspace"`
	At   [2]int `json:"at"`
	Size [2]int `json:"size"`
}

// LaunchCavaInTerminal launches Cava audio visualizer in a new terminal window
func LaunchCavaInTerminal() error {
	// First check if cava is available
	if err := checkCavaInstalled(); err != nil {
		return err
	}

	// If hyprctl is available, record current active window before launching
	var navAddr string
	var workspaceID int
	if _, err := exec.LookPath("hyprctl"); err == nil {
		if activeWin, err := getActiveHyprWindow(); err == nil {
			navAddr = activeWin.Address
			workspaceID = activeWin.Workspace.ID
		}
	}

	var err error
	switch runtime.GOOS {
	case "linux":
		err = launchLinuxTerminal()
	case "darwin":
		err = launchMacOSTerminal()
	case "windows":
		err = launchWindowsTerminal()
	default:
		err = launchFallbackTerminal()
	}

	if err != nil {
		return err
	}

	// If Hyprland is active, stack (cmd+j / togglesplit) and resize windows (80% / 20%)
	if navAddr != "" {
		go arrangeHyprlandCavaWindow(navAddr, workspaceID)
	}

	return nil
}

func getActiveHyprWindow() (*hyprWindow, error) {
	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		return nil, err
	}
	var win hyprWindow
	if err := json.Unmarshal(out, &win); err != nil {
		return nil, err
	}
	return &win, nil
}

func getHyprClients() ([]hyprWindow, error) {
	out, err := exec.Command("hyprctl", "clients", "-j").Output()
	if err != nil {
		return nil, err
	}
	var clients []hyprWindow
	if err := json.Unmarshal(out, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func arrangeHyprlandCavaWindow(navAddr string, workspaceID int) {
	var cavaAddr string
	// Poll up to 2 seconds for cava window to appear
	for i := 0; i < 40; i++ {
		time.Sleep(50 * time.Millisecond)
		clients, err := getHyprClients()
		if err != nil {
			continue
		}
		for _, w := range clients {
			if w.Workspace.ID == workspaceID && w.Address != navAddr {
				titleLower := strings.ToLower(w.Title)
				initTitleLower := strings.ToLower(w.InitialTitle)
				classLower := strings.ToLower(w.Class)
				if strings.Contains(titleLower, "cava") ||
					strings.Contains(initTitleLower, "cava") ||
					strings.Contains(classLower, "cava") ||
					w.Title != "" {
					cavaAddr = w.Address
					break
				}
			}
		}
		if cavaAddr != "" {
			break
		}
	}

	if cavaAddr == "" {
		return
	}

	// Give window system a moment to position new window
	time.Sleep(50 * time.Millisecond)

	clients, err := getHyprClients()
	if err != nil {
		return
	}

	var navWin, cavaWin *hyprWindow
	for i := range clients {
		if clients[i].Address == navAddr {
			navWin = &clients[i]
		} else if clients[i].Address == cavaAddr {
			cavaWin = &clients[i]
		}
	}

	if navWin == nil || cavaWin == nil {
		return
	}

	// 1. Perform cmd+j (layoutmsg togglesplit) if windows are side-by-side
	if navWin.At[0] != cavaWin.At[0] {
		exec.Command("hyprctl", "dispatch", "layoutmsg", "togglesplit").Run()
		time.Sleep(50 * time.Millisecond)

		if updated, err := getHyprClients(); err == nil {
			for i := range updated {
				if updated[i].Address == navAddr {
					navWin = &updated[i]
				} else if updated[i].Address == cavaAddr {
					cavaWin = &updated[i]
				}
			}
		}
	}

	// 2. Ensure Navitone is on top and Cava is on bottom
	if navWin.At[1] > cavaWin.At[1] {
		exec.Command("hyprctl", "dispatch", "focuswindow", fmt.Sprintf("address:%s", navAddr)).Run()
		exec.Command("hyprctl", "dispatch", "swapwindow", "u").Run()
		time.Sleep(50 * time.Millisecond)

		if updated, err := getHyprClients(); err == nil {
			for i := range updated {
				if updated[i].Address == navAddr {
					navWin = &updated[i]
				} else if updated[i].Address == cavaAddr {
					cavaWin = &updated[i]
				}
			}
		}
	}

	// 3. Resize Navitone height to 70% and Cava height to 30%
	totalH := navWin.Size[1] + cavaWin.Size[1]
	if totalH > 0 {
		targetNavH := int(float64(totalH) * 0.7)
		deltaY := targetNavH - navWin.Size[1]

		if deltaY != 0 {
			exec.Command("hyprctl", "dispatch", "focuswindow", fmt.Sprintf("address:%s", navAddr)).Run()
			exec.Command("hyprctl", "dispatch", "resizeactive", "0", fmt.Sprintf("%d", deltaY)).Run()
		}
	}

	// Keep focus on Navitone window
	exec.Command("hyprctl", "dispatch", "focuswindow", fmt.Sprintf("address:%s", navAddr)).Run()
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