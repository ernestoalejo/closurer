package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/ernestokarim/closurer/config"
)

var (
	modificationCache = map[string]time.Time{}
	sourcesCache      = map[string]*Source{}
	confsCache        = map[string]*config.Config{}
)

// Checks if filename has been modified since the last time
// it was scanned. It so, or if it's not present in the cache,
// it returns true and stores the new time.
func CacheModified(dest, filename string) (bool, error) {
	name := dest + filename

	if *noCache {
		return true, nil
	}

	info, err := os.Lstat(filename)
	if err != nil {
		return false, err
	}

	modified, ok := modificationCache[name]

	if !ok || info.ModTime() != modified {
		modificationCache[name] = info.ModTime()
		return true, nil
	}

	return false, nil
}

// Load the caches from a file.
func LoadCache() error {
	conf := config.Current()
	
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
	conf := config.Current()

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

func CachedConf(filename string) *config.Config {
	// Retrieve/Create a new entry in the configs map
	conf, ok := confsCache[filename]
	if !ok || *noCache {
		conf = new(config.Config)
		confsCache[filename] = conf
	}

	return conf
}

func CachedSource(dest, filename string) *Source {
	name := dest + filename

	source, ok := sourcesCache[name]
	if !ok || *noCache {
		source = new(Source)
		sourcesCache[name] = source
	}

	return source
}
