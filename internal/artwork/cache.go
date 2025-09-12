package artwork

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache handles artwork caching to improve performance
type Cache struct {
	cacheDir string
}

// NewCache creates a new artwork cache
func NewCache() (*Cache, error) {
	// Create cache directory in user's cache directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "navitone-cli", "artwork")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		cacheDir: cacheDir,
	}, nil
}

// Get retrieves cached ASCII art if it exists and is not expired
func (c *Cache) Get(key string) (string, bool) {
	filename := c.getCacheFilename(key)
	filepath := filepath.Join(c.cacheDir, filename)

	// Check if file exists and is not expired
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", false // File doesn't exist
	}

	// Check if file is expired (older than 30 days)
	if time.Since(stat.ModTime()) > 30*24*time.Hour {
		// Remove expired file
		os.Remove(filepath)
		return "", false
	}

	// Read cached content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", false
	}

	return string(content), true
}

// Set stores ASCII art in cache
func (c *Cache) Set(key, ascii string) error {
	filename := c.getCacheFilename(key)
	filepath := filepath.Join(c.cacheDir, filename)

	return os.WriteFile(filepath, []byte(ascii), 0644)
}

// getCacheFilename generates a safe filename from a cache key
func (c *Cache) getCacheFilename(key string) string {
	// Create MD5 hash of the key for safe filename
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x.txt", hash)
}

// Clean removes expired cache files
func (c *Cache) Clean() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	now := time.Now()
	maxAge := 30 * 24 * time.Hour

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .txt files (our cache files)
		if !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			filepath := filepath.Join(c.cacheDir, entry.Name())
			os.Remove(filepath)
		}
	}

	return nil
}

// GetCacheKey generates a cache key for an artwork URL
func GetCacheKey(url string, width, height int) string {
	return fmt.Sprintf("%s_%dx%d", url, width, height)
}