package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type Config struct {
	Id string `json:"id"`

	// Root folder where the code can be found
	RootJs  string `json:"root-js"`
	RootSoy string `json:"root-soy"`
	RootGss string `json:"root-gss"`

	// Extern scripts folder
	Externs []string `json:"externs"`

	// Temporary build directory
	Build string `json:"build"`

	// Closure utilities paths
	ClosureLibrary     string `json:"closure-library"`
	ClosureCompiler    string `json:"closure-compiler"`
	ClosureTemplates   string `json:"closure-templates"`
	ClosureStylesheets string `json:"closure-stylesheets"`

	// Compilation mode: SIMPLE, ADVANCE, WHITESPACE
	Mode string `json:"mode"`

	// Warnings level: QUIET, DEFAULT, VERBOSE
	Level string `json:"level"`

	// List of inputs (main files where the compilation starts)
	Inputs []string `json:"inputs"`

	// Defines if each check emits a WARNING, an ERROR or it's OFF.
	Checks map[string]string `json:"checks"`

	// Define additional values in the compilation
	Define map[string]string `json:"define"`

	// Inherits another configurations file
	Inherits string `json:"inherits"`
}

var conf = new(Config)
var confs = map[string]*Config{}
var confModified = map[string]time.Time{}

func ReadConf() error {
	if err := LoadConfFile(*confArg); err != nil {
		return err
	}

	return nil
}

func LoadConfFile(filename string) error {
	config, ok := confs[filename]
	if !ok {
		config = new(Config)
	}

	// Check the modified time
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	modified, ok := confModified[filename]
	if !ok || info.ModTime() != modified {
		confModified[filename] = info.ModTime()

		log.Println("Reading config file:", filename)

		// Open the file
		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("cannot open the config file %s: %s", filename, err)
		}
		defer f.Close()

		// Load the data
		dec := json.NewDecoder(f)
		if err := dec.Decode(config); err != nil {
			return err
		}

		// Adjust the path if necessary
		if config.Inherits != "" {
			config.Inherits = path.Join(path.Dir(filename), config.Inherits)
		}

		confs[filename] = config

		// Invalid caches
		sourcesCache = map[string]*Source{}
		soyCache = map[string]time.Time{}
		gssCache = map[string]time.Time{}
	}

	if config.Inherits != "" {
		if err := LoadConfFile(config.Inherits); err != nil {
			return err
		}
	}

	ApplyConf(config)

	return nil
}

func ApplyConf(config *Config) {
	if config.Id != "" {
		conf.Id = config.Id
	}

	if config.RootJs != "" {
		conf.RootJs = config.RootJs
	}
	if config.RootGss != "" {
		conf.RootGss = config.RootGss
	}
	if config.RootSoy != "" {
		conf.RootSoy = config.RootSoy
	}
	if len(config.Externs) != 0 {
		conf.Externs = config.Externs
	}
	if config.Build != "" {
		conf.Build = config.Build
	}

	if config.ClosureLibrary != "" {
		conf.ClosureLibrary = config.ClosureLibrary
	}
	if config.ClosureCompiler != "" {
		conf.ClosureCompiler = config.ClosureCompiler
	}
	if config.ClosureTemplates != "" {
		conf.ClosureTemplates = config.ClosureTemplates
	}
	if config.ClosureStylesheets != "" {
		conf.ClosureStylesheets = config.ClosureStylesheets
	}

	if config.Mode != "" {
		conf.Mode = config.Mode
	}
	if config.Level != "" {
		conf.Level = config.Level
	}

	if len(config.Inputs) > 0 {
		conf.Inputs = config.Inputs
	}
	if len(config.Checks) > 0 {
		conf.Checks = config.Checks
	}

	if len(config.Define) > 0 {
		conf.Define = config.Define
	}
}
