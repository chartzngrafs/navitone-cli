package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// IPCClient handles JSON IPC communication with MPV
type IPCClient struct {
	conn     net.Conn
	requests map[int]chan IPCResponse
	events   chan MPVEvent
	
	// Request ID counter
	requestID int64
	
	// State
	connected bool
	
	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup
}

// MPVCommand represents a command sent to MPV
type MPVCommand struct {
	Command   []interface{} `json:"command"`
	RequestID int           `json:"request_id,omitempty"`
}

// IPCResponse represents a response from MPV
type IPCResponse struct {
	RequestID int         `json:"request_id"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

// MPVEvent represents an event from MPV
type MPVEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:",inline"`
	
	// Common event properties
	Reason   string  `json:"reason,omitempty"`
	Filename string  `json:"filename,omitempty"`
	Path     string  `json:"path,omitempty"`
	Position float64 `json:"playback-time,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Pause    bool    `json:"pause,omitempty"`
	EOF      bool    `json:"eof-reached,omitempty"`
}

// NewIPCClient creates a new IPC client for the given socket path
func NewIPCClient(socketPath string) (*IPCClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MPV socket: %w", err)
	}

	client := &IPCClient{
		conn:      conn,
		requests:  make(map[int]chan IPCResponse),
		events:    make(chan MPVEvent, 100), // Buffered channel for events
		connected: true,
	}

	// Start reader goroutine
	client.wg.Add(1)
	go client.readLoop()

	return client, nil
}

// SendCommand sends a command to MPV and waits for response
func (c *IPCClient) SendCommand(command string, args ...interface{}) (interface{}, error) {
	return c.SendCommandWithTimeout(5*time.Second, command, args...)
}

// SendCommandWithTimeout sends a command with a custom timeout
func (c *IPCClient) SendCommandWithTimeout(timeout time.Duration, command string, args ...interface{}) (interface{}, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("IPC client not connected")
	}

	// Build command array
	cmdArray := make([]interface{}, 1+len(args))
	cmdArray[0] = command
	copy(cmdArray[1:], args)

	// Generate request ID
	reqID := int(atomic.AddInt64(&c.requestID, 1))

	// Create response channel
	respChan := make(chan IPCResponse, 1)
	
	c.mu.Lock()
	c.requests[reqID] = respChan
	c.mu.Unlock()

	// Clean up on exit
	defer func() {
		c.mu.Lock()
		delete(c.requests, reqID)
		close(respChan)
		c.mu.Unlock()
	}()

	// Send command
	cmd := MPVCommand{
		Command:   cmdArray,
		RequestID: reqID,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	data = append(data, '\n')

	c.mu.Lock()
	_, err = c.conn.Write(data)
	c.mu.Unlock()
	
	if err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		if resp.Error != "" && resp.Error != "success" {
			return nil, fmt.Errorf("MPV error: %s", resp.Error)
		}
		return resp.Data, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("command timeout after %v", timeout)
	}
}

// SendCommandAsync sends a command without waiting for response
func (c *IPCClient) SendCommandAsync(command string, args ...interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("IPC client not connected")
	}

	// Build command array
	cmdArray := make([]interface{}, 1+len(args))
	cmdArray[0] = command
	copy(cmdArray[1:], args)

	// Send without request ID (fire and forget)
	cmd := MPVCommand{
		Command: cmdArray,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	data = append(data, '\n')

	c.mu.Lock()
	_, err = c.conn.Write(data)
	c.mu.Unlock()
	
	return err
}

// GetEvents returns the events channel
func (c *IPCClient) GetEvents() <-chan MPVEvent {
	return c.events
}

// IsConnected returns whether the client is connected
func (c *IPCClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Close closes the IPC connection
func (c *IPCClient) Close() error {
	c.mu.Lock()
	if c.connected {
		c.connected = false
		c.conn.Close()
	}
	c.mu.Unlock()

	c.wg.Wait()
	
	// Close events channel
	close(c.events)
	
	return nil
}

// Reconnect attempts to reconnect to the socket
func (c *IPCClient) Reconnect(socketPath string) error {
	c.Close()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to reconnect to MPV socket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.requests = make(map[int]chan IPCResponse)
	c.events = make(chan MPVEvent, 100)
	c.mu.Unlock()

	// Restart reader goroutine
	c.wg.Add(1)
	go c.readLoop()

	return nil
}

// readLoop handles reading responses and events from MPV
func (c *IPCClient) readLoop() {
	defer c.wg.Done()
	
	scanner := bufio.NewScanner(c.conn)
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Try to parse as response first
		var resp IPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err == nil && resp.RequestID != 0 {
			// This is a response to a command
			c.mu.RLock()
			if respChan, exists := c.requests[resp.RequestID]; exists {
				select {
				case respChan <- resp:
				default:
					// Channel full or closed, ignore
				}
			}
			c.mu.RUnlock()
			continue
		}

		// Try to parse as event
		var event MPVEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil && event.Event != "" {
			// This is an event
			select {
			case c.events <- event:
			default:
				// Events channel full, drop event
			}
			continue
		}

		// Unknown message format, ignore
	}

	// Scanner finished (connection closed)
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}