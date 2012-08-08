package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

var (
	modificationCache = map[string]time.Time{}
	sourcesCache      = map[string]*Source{}
	confsCache        = map[string]*Config{}
)

// Checks if filename has been modified since the last time
// it was scanned. It so, or if it's not present in the cache,
// it returns true and stores the new time.
func CacheModified(filename string) (bool, error) {
	if *noCache {
		return true, nil
	}

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
	if *noCache {
		return nil
	}

	name := path.Join(conf.Build, "cache")

	// Open the cache file if it exists
	f, err := os.Open(name)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot open the cache file: %s", err)
	}
	defer f.Close()

	log.Println("Reading deps cache:", name)

	// Decode the caches
	d := gob.NewDecoder(f)

	if err := d.Decode(&modificationCache); err != nil {
		return fmt.Errorf("cannot decode the modifications cache: %s", err)
	}
	if err := d.Decode(&sourcesCache); err != nil {
		return fmt.Errorf("cannot decode the sources cache: %s", err)
	}
	if err := d.Decode(&confsCache); err != nil {
		return fmt.Errorf("cannot decode the confs cache: %s", err)
	}

	return nil
}

// Save the caches to a file.
func WriteCache() error {
	if *noCache {
		return nil
	}

	// Create the cache file
	f, err := os.Create(path.Join(conf.Build, "cache"))
	if err != nil {
		return fmt.Errorf("cannot create the cache file: %s", err)
	}
	defer f.Close()

	// Encode the caches
	e := gob.NewEncoder(f)

	if err := e.Encode(&modificationCache); err != nil {
		return fmt.Errorf("cannot encode the modifications cache: %s", err)
	}
	if err := e.Encode(&sourcesCache); err != nil {
		return fmt.Errorf("cannot encode the sources cache: %s", err)
	}
	if err := e.Encode(&confsCache); err != nil {
		return fmt.Errorf("cannot encode the confs cache: %s", err)
	}

	return nil
}

func CachedConf(filename string) *Config {
	// Retrieve/Create a new entry in the configs map
	config, ok := confsCache[filename]
	if !ok || *noCache {
		config = new(Config)
		confsCache[filename] = config
	}

	return config
}

func CachedSource(filename string) *Source {
	source, ok := sourcesCache[filename]
	if !ok || *noCache {
		source = new(Source)
		sourcesCache[filename] = source
	}

	return source
}
