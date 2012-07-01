package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
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

// Load the caches from a file.
func LoadCache() error {
	name := path.Join(conf.Build, "cache")

	// Open the cache file if it exists
	f, err := os.Open(name)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	log.Println("Reading deps cache:", name)

	// Decode the caches
	if err := gob.NewDecoder(f).Decode(&modificationCache); err != nil {
		return fmt.Errorf("cannot decode the deps cache: %s", err)
	}

	return nil
}

// Save the caches to a file.
func WriteCache() error {
	// Create the cache file
	f, err := os.Create(path.Join(conf.Build, "cache"))
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode the caches
	if err := gob.NewEncoder(f).Encode(&modificationCache); err != nil {
		return fmt.Errorf("cannot encode the deps cache: %s", err)
	}

	return nil
}
