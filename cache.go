package main

import (
	"os"
	"time"
)

var modificationCache = map[string]time.Time{}

// Checks if filename has been modified since the last time
// it was scanned. It so, or if it's not present in the cache,
// it returns true and stores the new time.
func CacheModified(filename string) (bool, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return false, err
	}

	modified, ok := modificationCache[filename]

	if !ok || info.ModTime() != modified {
		modificationCache[filename] = info.ModTime()
		return true, nil
	}

	return false, nil
}
