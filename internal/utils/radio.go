package utils

import (
	"net/url"
	"path/filepath"
	"strings"
)

// NormalizeRadioURL fixes known deprecated, non-standard, un-extensioned, or whitespace-padded stream URL formats for internet radio (e.g. Nightride FM)
func NormalizeRadioURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	if !strings.Contains(strings.ToLower(rawURL), "nightride.fm") {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		if strings.HasSuffix(rawURL, ".m4a") {
			return strings.TrimSuffix(rawURL, ".m4a") + ".mp3"
		}
		return rawURL
	}

	p := u.Path
	if p == "" || p == "/" {
		return u.String()
	}

	pLower := strings.ToLower(p)
	if strings.HasSuffix(pLower, ".mp3") || strings.HasSuffix(pLower, ".ogg") {
		return u.String()
	}

	// Remove extension if present (e.g. .m4a, .opus, .aac, .flac)
	ext := filepath.Ext(p)
	if ext != "" {
		p = p[:len(p)-len(ext)]
	}
	p = strings.TrimSuffix(p, "/")

	if p != "" {
		u.Path = p + ".mp3"
		return u.String()
	}

	return u.String()
}
