package main

import (
	"sync"
)

var loadCacheOnce sync.Once

// Called before each compilation task. It load the caches
// and reload the confs if needed.
func PreCompileActions() error {
	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	// Load the cache the first time is needed
	var err error
	loadCacheOnce.Do(func() {
		err = LoadCache()
	})

	return err
}

// Called after each compilation tasks. It saves the caches.
func PostCompileActions() error {
	return WriteCache()
}
