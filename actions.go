package main

import (
)

// Called before each compilation task. It load the caches
// and reload the confs if needed.
func PreCompileActions() error {
	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	return nil
}

// Called after each compilation tasks. It saves the caches.
func PostCompileActions() error {
	return nil
}
