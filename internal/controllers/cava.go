package controllers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	cavaCmd   *exec.Cmd
	cavaMu    sync.Mutex
	cavaActive bool
)

// toggleInAppCava toggles the integrated in-app Cava visualizer
func (a *App) toggleInAppCava() {
	a.state.Mu.Lock()
	a.state.ShowCavaVisualizer = !a.state.ShowCavaVisualizer
	active := a.state.ShowCavaVisualizer
	a.state.Mu.Unlock()

	if active {
		a.logMessage("In-app Cava visualizer enabled")
		a.startCavaPipe()
	} else {
		a.logMessage("In-app Cava visualizer disabled")
		a.stopCavaPipe()
	}
}

// startCavaPipe starts background cava raw ASCII stdout reader if available
func (a *App) startCavaPipe() {
	cavaMu.Lock()
	defer cavaMu.Unlock()

	if cavaActive {
		return
	}

	cavaPath, err := exec.LookPath("cava")
	if err != nil {
		a.logMessage("Cava binary not found; using built-in dynamic spectrum backend")
		cavaActive = true
		return
	}

	// Create temporary config file for raw ASCII stdout output
	tmpDir := os.TempDir()
	cfgPath := filepath.Join(tmpDir, "navitone_cava_raw.conf")
	cfgContent := `[general]
framerate = 120
sensitivity = 165
autosens = 1
bars = 0
bar_width = 2
bar_spacing = 1
lower_cutoff_freq = 40
higher_cutoff_freq = 20000

[input]
method = pulse
source = auto

[output]
method = raw
raw_target = /dev/stdout
data_format = ascii
ascii_max_range = 1000
bar_delimiter = 59
channels = mono
mono_option = average

[smoothing]
noise_reduction = 80
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		a.logMessage(fmt.Sprintf("Failed to write Cava temp config: %v", err))
		cavaActive = true
		return
	}

	cmd := exec.Command(cavaPath, "-p", cfgPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.logMessage(fmt.Sprintf("Failed to pipe Cava stdout: %v", err))
		cavaActive = true
		return
	}

	if err := cmd.Start(); err != nil {
		a.logMessage(fmt.Sprintf("Failed to start Cava pipe: %v", err))
		cavaActive = true
		return
	}

	cavaCmd = cmd
	cavaActive = true

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			cavaMu.Lock()
			if !cavaActive {
				cavaMu.Unlock()
				break
			}
			cavaMu.Unlock()

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			parts := strings.Split(line, ";")
			var bars []int
			for _, p := range parts {
				if p == "" {
					continue
				}
				if val, err := strconv.Atoi(p); err == nil {
					bars = append(bars, val)
				}
			}

			if len(bars) > 0 {
				a.state.Mu.Lock()
				a.state.CavaBars = bars
				a.state.Mu.Unlock()
			}
		}

		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()
}

// stopCavaPipe stops background cava process
func (a *App) stopCavaPipe() {
	cavaMu.Lock()
	defer cavaMu.Unlock()

	cavaActive = false
	if cavaCmd != nil && cavaCmd.Process != nil {
		_ = cavaCmd.Process.Kill()
		cavaCmd = nil
	}
}
