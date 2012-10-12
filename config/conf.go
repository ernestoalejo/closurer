package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ernestokarim/closurer/cache"
)

type Config struct {
	Id string `json:"id"`

	// Root folder where the code can be found
	RootJs  string `json:"root-js"`
	RootSoy string `json:"root-soy"`
	RootGss string `json:"root-gss"`

	// Extern scripts folders
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

// Global configuration.
var conf = new(Config)

// Load a config file recursively (inheritation) and apply
// the settings to the global object.
func ReadFromFile(filename string) error {
	config := cache.ReadData(filename, new(Config)).(*Config)

	// Check the modified time
	if modified, err := cache.Modified("config", filename); err != nil {
		return err
	} else if modified {
		log.Println("Reading config file:", filename)

		// Open the file
		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("cannot open the config file %s: %s", filename, err)
		}
		defer f.Close()

		// Load the data
		if err := json.NewDecoder(f).Decode(config); err != nil {
			return err
		}

		// Adjust the paths
		config.ClosureLibrary = fixPath(config.ClosureLibrary)
		config.ClosureCompiler = fixPath(config.ClosureCompiler)
		config.ClosureTemplates = fixPath(config.ClosureTemplates)
		config.ClosureStylesheets = fixPath(config.ClosureStylesheets)
		config.Inherits = fixInheritsPath(filename, config.Inherits)
	}

	// Recursively scan inherited files
	if config.Inherits != "" {
		if err := ReadFromFile(config.Inherits); err != nil {
			return err
		}
	}

	applyConf(config)

	return nil
}

// Copy the non-zero settings from config to the global
// conf object.
func applyConf(config *Config) {
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
	if len(config.Externs) > 0 {
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

// Replace the ~ with the correct folder path
func fixPath(p string) string {
	if !strings.Contains(p, "~") {
		return p
	}

	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		log.Fatal("found ~ in a path, but USER nor USERNAME are setted in the env")
	}

	return strings.Replace(p, "~", "/home/"+user, -1)
}

// Converts a relative path to current, to an absolute path
// if needed.
func fixInheritsPath(current string, p string) string {
	if p != "" && !path.IsAbs(p) {
		p = path.Join(path.Dir(current), p)
	}
	return p
}

func Current() *Config {
	return conf
}