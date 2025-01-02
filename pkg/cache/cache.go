package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache provides a simple file-based caching mechanism
type Cache struct {
	dir string
	mu  sync.RWMutex
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Value      interface{} `json:"value"`
	Timestamp  time.Time   `json:"timestamp"`
	Version    string      `json:"version"`
	ExpiresAt  time.Time   `json:"expires_at"`
}

// NewCache creates a new cache instance
func NewCache(dir string) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &Cache{dir: dir}, nil
}

// generateKey creates a unique key for the cache entry
func generateKey(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		// If marshaling fails, use string representation
		return fmt.Sprintf("%v", data)
	}
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// Get retrieves a value from the cache
func (c *Cache) Get(key interface{}, result interface{}) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheKey := generateKey(key)
	path := filepath.Join(c.dir, cacheKey+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if cache entry has expired
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		os.Remove(path) // Clean up expired entry
		return false, nil
	}

	// Unmarshal the cached value into the result
	valueBytes, err := json.Marshal(entry.Value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cached value: %w", err)
	}

	if err := json.Unmarshal(valueBytes, result); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return true, nil
}

// Set stores a value in the cache
func (c *Cache) Set(key interface{}, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheKey := generateKey(key)
	path := filepath.Join(c.dir, cacheKey+".json")

	entry := CacheEntry{
		Value:     value,
		Timestamp: time.Now(),
		Version:   "1.0", // Version can be used for cache invalidation
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all entries from the cache
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir, err := os.Open(c.dir)
	if err != nil {
		return fmt.Errorf("failed to open cache directory: %w", err)
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, name := range names {
		path := filepath.Join(c.dir, name)
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove cache file %s: %w", name, err)
		}
	}

	return nil
}

// Cleanup removes expired entries from the cache
func (c *Cache) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir, err := os.Open(c.dir)
	if err != nil {
		return fmt.Errorf("failed to open cache directory: %w", err)
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, name := range names {
		path := filepath.Join(c.dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files we can't read
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue // Skip invalid entries
		}

		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			os.Remove(path)
		}
	}

	return nil
}
