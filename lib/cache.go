package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type CacheEntry struct {
	Updated time.Time      `json:"updated"`
	Status  *ServiceStatus `json:"status"`
}

type Cache struct {
	path    string
	entries map[string]CacheEntry
}

func NewCache() (*Cache, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	return &Cache{
		path:    filepath.Join(dir, "cloudcash", "status.json"),
		entries: make(map[string]CacheEntry),
	}, nil
}

func (c *Cache) Load() error {
	raw, err := os.ReadFile(c.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	entries := make(map[string]CacheEntry)
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("%s: %w", c.path, err)
	}

	c.entries = entries
	return nil
}

func (c *Cache) Get(id string) (CacheEntry, bool) {
	entry, ok := c.entries[id]
	return entry, ok && entry.Status != nil
}

func (c *Cache) Put(id string, entry CacheEntry) {
	c.entries[id] = entry
}

func (c *Cache) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".status-*.json")
	if err != nil {
		return err
	}

	err = json.NewEncoder(tmp).Encode(c.entries)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), c.path)
	}
	if err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}

	return nil
}
