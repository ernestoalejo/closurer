package cache

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

// This flag will be loaded by the config package; it must be here to
// avoid a circular dependency.
var NoCache bool

// Load the caches from a file.
func Load(filename string) error {
	if NoCache {
		return nil
	}

	f, err := os.Open(filename)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot open the cache file: %s", err)
	}
	defer f.Close()

	log.Println("Reading deps cache:", filename)

	d := gob.NewDecoder(f)
	if err := d.Decode(&modificationCache); err != nil {
		return fmt.Errorf("cannot decode the modifications cache: %s", err)
	}
	if err := d.Decode(&dataCache); err != nil {
		return fmt.Errorf("cannot decode the confs cache: %s", err)
	}

	return nil
}

// Save the caches to a file.
func Dump(filename string) error {
	// Create the cache file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create the cache file: %s", err)
	}
	defer f.Close()

	// Encode the caches
	e := gob.NewEncoder(f)

	if err := e.Encode(&modificationCache); err != nil {
		return fmt.Errorf("cannot encode the modifications cache: %s", err)
	}
	if err := e.Encode(&dataCache); err != nil {
		return fmt.Errorf("cannot encode the confs cache: %s", err)
	}

	return nil
}
