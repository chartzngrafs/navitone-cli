package mpv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// MPVProcess manages the MPV subprocess
type MPVProcess struct {
	process    *exec.Cmd
	socketPath string
	logPath    string
	ipc        *IPCClient

	// State
	isRunning bool
	mu        sync.RWMutex
}

// NewMPVProcess creates a new MPV process manager
func NewMPVProcess(socketPath string) *MPVProcess {
	if socketPath == "" {
		socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("navitone_mpv_%d", time.Now().Unix()))
	}

	return &MPVProcess{
		socketPath: socketPath,
		logPath:    socketPath + ".log",
	}
}

// Start starts the MPV process with the given arguments
func (m *MPVProcess) Start(args []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("MPV process is already running")
	}

	// Check if MPV is available in PATH
	if _, err := exec.LookPath("mpv"); err != nil {
		return fmt.Errorf("mpv binary not found in PATH: %w", err)
	}

	// Remove old socket file if it exists
	os.Remove(m.socketPath)

	// Default MPV arguments for audio-only playback
	defaultArgs := []string{
		"--no-video",              // Audio only
		"--idle",                  // Stay alive when no file is playing
		"--no-terminal",           // Don't use terminal for input
		"--msg-level=all=error",   // Reduce log verbosity
		"--audio-buffer=0.5",      // 500ms audio buffer
		"--gapless-audio=yes",     // Enable gapless playback
		"--replaygain=track",      // Enable replay gain
		"--volume=70",             // Default volume 70%
		fmt.Sprintf("--input-ipc-server=%s", m.socketPath), // IPC socket
		fmt.Sprintf("--log-file=%s", m.logPath),             // Log file
	}

	// Combine default args with user args
	allArgs := append(defaultArgs, args...)

	// Create the command
	m.process = exec.Command("mpv", allArgs...)

	// Set process group to allow clean termination
	m.process.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Redirect stdout and stderr to prevent MPV output from interfering with TUI
	m.process.Stdout = nil
	m.process.Stderr = nil

	// Start the process
	if err := m.process.Start(); err != nil {
		return fmt.Errorf("failed to start MPV process: %w", err)
	}

	m.isRunning = true

	// Wait a bit for MPV to create the socket
	time.Sleep(500 * time.Millisecond)

	// Create IPC client
	ipc, err := NewIPCClient(m.socketPath)
	if err != nil {
		m.Stop() // Clean up
		return fmt.Errorf("failed to create IPC client: %w", err)
	}

	m.ipc = ipc

	// Start monitoring the process
	go m.monitorProcess()

	return nil
}

// Stop stops the MPV process
func (m *MPVProcess) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning || m.process == nil {
		return nil
	}

	// Close IPC connection first
	if m.ipc != nil {
		m.ipc.Close()
		m.ipc = nil
	}

	// Try graceful shutdown first
	if err := m.process.Process.Signal(syscall.SIGTERM); err == nil {
		// Wait up to 5 seconds for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- m.process.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(5 * time.Second):
			// Force kill after timeout
			m.process.Process.Kill()
			// Also kill process group to ensure cleanup of any child processes
			if m.process.Process.Pid > 0 {
				syscall.Kill(-m.process.Process.Pid, syscall.SIGKILL)
			}
			m.process.Wait()
		}
	} else {
		// Immediate kill if signal failed
		m.process.Process.Kill()
		// Also kill process group to ensure cleanup of any child processes
		if m.process.Process.Pid > 0 {
			syscall.Kill(-m.process.Process.Pid, syscall.SIGKILL)
		}
		m.process.Wait()
	}

	m.isRunning = false
	m.process = nil

	// Clean up files
	os.Remove(m.socketPath)
	os.Remove(m.logPath)

	return nil
}

// IsRunning returns whether the MPV process is running
func (m *MPVProcess) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning && m.process != nil
}

// GetIPC returns the IPC client for communication with MPV
func (m *MPVProcess) GetIPC() *IPCClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ipc
}

// GetSocketPath returns the path to the IPC socket
func (m *MPVProcess) GetSocketPath() string {
	return m.socketPath
}

// GetLogPath returns the path to the log file
func (m *MPVProcess) GetLogPath() string {
	return m.logPath
}

// monitorProcess monitors the MPV process and handles unexpected exits
func (m *MPVProcess) monitorProcess() {
	if m.process == nil {
		return
	}

	// Wait for process to exit
	err := m.process.Wait()

	m.mu.Lock()
	m.isRunning = false
	if m.ipc != nil {
		m.ipc.Close()
		m.ipc = nil
	}
	m.mu.Unlock()

	if err != nil {
		// Process exited with error - this could be logged or handled
		// For now, we'll just mark it as not running
	}

	// Clean up files
	os.Remove(m.socketPath)
	os.Remove(m.logPath)
}

// Restart restarts the MPV process
func (m *MPVProcess) Restart(args []string) error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop existing process: %w", err)
	}

	// Wait a bit before restarting
	time.Sleep(500 * time.Millisecond)

	return m.Start(args)
}