package cache

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
)

// Load the caches from a file.
func Load(filename string) error {
	if config.NoCache {
		return nil
	}

	f, err := os.Open(filename)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	log.Println("Reading deps cache:", filename)

	d := gob.NewDecoder(f)
	if err := d.Decode(&modificationCache); err != nil {
		return app.Error(err)
	}
	if err := d.Decode(&dataCache); err != nil {
		return app.Error(err)
	}

	return nil
}

// Save the caches to a file.
func Dump(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	e := gob.NewEncoder(f)
	if err := e.Encode(&modificationCache); err != nil {
		return app.Error(err)
	}
	if err := e.Encode(&dataCache); err != nil {
		return app.Error(err)
	}

	return nil
}
