package artwork

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}

	if cache == nil {
		t.Fatalf("NewCache returned nil")
	}

	if cache.cacheDir == "" {
		t.Errorf("Cache directory path is empty")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	// Create a temporary cache directory
	tmpDir, err := os.MkdirTemp("", "navitone-artwork-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cache := &Cache{cacheDir: tmpDir}
	key := "album-12345"
	artContent := "ASCII ART CONTENT HERE"

	// Check missing key
	_, found := cache.Get(key)
	if found {
		t.Errorf("Expected key not to be found before Set")
	}

	// Set key
	err = cache.Set(key, artContent)
	if err != nil {
		t.Fatalf("Cache Set failed: %v", err)
	}

	// Get key
	retrieved, found := cache.Get(key)
	if !found {
		t.Fatalf("Expected key to be found after Set")
	}

	if retrieved != artContent {
		t.Errorf("Expected '%s', got '%s'", artContent, retrieved)
	}
}

func TestGetCacheFilename(t *testing.T) {
	cache := &Cache{cacheDir: "/tmp"}
	filename1 := cache.getCacheFilename("test-key-1")
	filename2 := cache.getCacheFilename("test-key-2")

	if filename1 == filename2 {
		t.Errorf("Expected different filenames for different keys")
	}

	if filepath.Ext(filename1) != ".txt" {
		t.Errorf("Expected .txt extension, got '%s'", filepath.Ext(filename1))
	}
}
