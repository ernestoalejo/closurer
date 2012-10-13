package config

import (
	"os"
	"time"
)

var (
	configCache       = map[string]*Config{}
	modificationCache = map[string]time.Time{}
)

func cacheReadConfig(key string) *Config {
	d, ok := configCache[key]
	if !ok || NoCache {
		configCache[key] = new(Config)
		return configCache[key]
	}

	return d
}

func cacheModified(filename string) (bool, error) {
	if NoCache {
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
