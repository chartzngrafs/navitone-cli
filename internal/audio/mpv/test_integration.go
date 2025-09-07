package mpv

import (
	"fmt"
	"time"
)

// TestMPVIntegration tests basic MPV functionality
func TestMPVIntegration() error {
	fmt.Println("Testing MPV integration...")

	// Test MPV process creation
	process := NewMPVProcess("")
	if process == nil {
		return fmt.Errorf("failed to create MPV process")
	}

	// Test starting MPV
	fmt.Println("Starting MPV process...")
	err := process.Start(nil)
	if err != nil {
		return fmt.Errorf("failed to start MPV: %v", err)
	}

	// Give MPV time to start
	time.Sleep(1 * time.Second)

	// Test IPC connection
	ipc := process.GetIPC()
	if ipc == nil {
		process.Stop()
		return fmt.Errorf("failed to get IPC client")
	}

	fmt.Println("Testing IPC connection...")
	if !ipc.IsConnected() {
		process.Stop()
		return fmt.Errorf("IPC not connected")
	}

	// Test basic command
	commands := NewCommandWrapper(ipc)
	if commands == nil {
		process.Stop()
		return fmt.Errorf("failed to create command wrapper")
	}

	// Test getting version
	fmt.Println("Getting MPV version...")
	version, err := commands.GetVersion()
	if err != nil {
		process.Stop()
		return fmt.Errorf("failed to get MPV version: %v", err)
	}
	fmt.Printf("MPV Version: %s\n", version)

	// Test volume control
	fmt.Println("Testing volume control...")
	err = commands.SetVolume(50)
	if err != nil {
		process.Stop()
		return fmt.Errorf("failed to set volume: %v", err)
	}

	volume, err := commands.GetVolume()
	if err != nil {
		process.Stop()
		return fmt.Errorf("failed to get volume: %v", err)
	}
	fmt.Printf("Volume: %.0f%%\n", volume)

	// Clean up
	fmt.Println("Stopping MPV process...")
	err = process.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop MPV: %v", err)
	}

	fmt.Println("MPV integration test completed successfully!")
	return nil
}

// RunIntegrationTest is a helper function to run the integration test
func RunIntegrationTest() {
	fmt.Println("=== MPV Integration Test ===")
	if err := TestMPVIntegration(); err != nil {
		fmt.Printf("Integration test failed: %v\n", err)
	} else {
		fmt.Println("Integration test passed!")
	}
	fmt.Println("============================")
}