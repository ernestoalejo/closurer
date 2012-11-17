package cache

import (
	"encoding/gob"
	"log"
	"os"
	"path/filepath"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
)

const CACHE_FILENAME = "cache"

// Load the caches from a file.
func Load() error {
	conf := config.Current()
	filename := filepath.Join(conf.Build, CACHE_FILENAME)

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

	log.Println("Reading cache:", filename)

	d := gob.NewDecoder(f)
	if err := d.Decode(&modificationCache); err != nil {
		return app.Error(err)
	}
	if err := d.Decode(&dataCache); err != nil {
		return app.Error(err)
	}

	log.Println("Read", len(modificationCache), "modifications and", len(dataCache), "datas!")

	return nil
}

// Save the caches to a file.
func Dump() error {
	conf := config.Current()

	f, err := os.Create(filepath.Join(conf.Build, CACHE_FILENAME))
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	log.Println("Write", len(modificationCache), "modifications and", len(dataCache), "datas!")

	e := gob.NewEncoder(f)
	if err := e.Encode(&modificationCache); err != nil {
		return app.Error(err)
	}
	if err := e.Encode(&dataCache); err != nil {
		return app.Error(err)
	}

	return nil
}
