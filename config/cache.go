package config

import (
	"os"
	"time"
)

var (
	configCache       = map[string]*Config{}
	modificationCache = map[string]time.Time{}
)

func CacheReadConfig(key string) *Config {
	d, ok := configCache[key]
	if !ok || NoCache {
		configCache[key] = new(Config)
		return configCache[key]
	}

	return d
}

func CacheModified(dest, filename string) (bool, error) {
	if NoCache {
		return true, nil
	}

	name := dest + filename

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
