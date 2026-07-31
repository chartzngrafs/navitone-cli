package navidrome

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://music.example.com", "user", "pass")
	if client == nil {
		t.Fatalf("NewClient returned nil")
	}

	if client.baseURL != "https://music.example.com" {
		t.Errorf("Expected baseURL 'https://music.example.com', got '%s'", client.baseURL)
	}

	if client.username != "user" {
		t.Errorf("Expected username 'user', got '%s'", client.username)
	}
}

func TestNormalizeURL(t *testing.T) {
	client := NewClient("https://music.example.com/", "user", "pass")
	if client.baseURL != "https://music.example.com" {
		t.Errorf("Expected trailing slash to be trimmed from server URL, got '%s'", client.baseURL)
	}
}
