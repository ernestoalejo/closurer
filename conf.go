package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Config struct {
	Id string `json:"id"`

	Root  string   `json:"root"`
	Paths []string `json:"paths"`
	Build string   `json:"build"`

	ClosureLibrary   string `json:"closure-library"`
	ClosureCompiler  string `json:"closure-compiler"`
	ClosureTemplates string `json:"closure-templates"`

	Mode  string `json:"mode"`
	Level string `json:"level"`

	Inputs []string `json:"inputs"`

	Checks map[string]string `json:"checks"`
	Define map[string]string `json:"define"`
}

var conf = new(Config)
var confModified time.Time

func ReadConf() error {
	info, err := os.Lstat(*confArg)
	if err != nil {
		return err
	}

	if !confModified.IsZero() && info.ModTime() == confModified {
		return nil
	}
	confModified = info.ModTime()

	f, err := os.Open(*confArg)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(conf); err != nil {
		return err
	}

	log.Println("Read app config: ", conf.Id)

	// Invalid caches
	sourcesCache = map[string]*Source{}
	timesCache = map[string]time.Time{}

	return nil
}
