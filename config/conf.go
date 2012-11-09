package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
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

	// Compilation mode: RAW, SIMPLE, ADVANCE, WHITESPACE
	Mode string `json:"mode"`

	// Warnings level: QUIET, DEFAULT, VERBOSE
	Level string `json:"level"`

	// List of inputs (main files where the compilation starts)
	Inputs []string `json:"inputs"`

	// Defines if each check emits a WARNING, an ERROR or it's OFF.
	Checks map[string]string `json:"checks"`

	// Define additional values in the compilation
	Define map[string]string `json:"define"`

	// Define additional non-standard functions that can be used
	// in the .gss files.
	NonStandardCssFuncs []string `json:"non-standard-css-funcs"`

	// Rename the CSS classes to a shorter form
	RenameCss string `json:"rename-css"`

	// Inherits another configurations file
	Inherits string `json:"inherits"`
}

// Global configuration.
var conf = new(Config)

func Current() *Config {
	return conf
}

// Load a config file recursively (inheritation) and apply
// the settings to the global object.
func ReadFromFile(filename string) error {
	config := cacheReadConfig(filename)

	// Check the modified time
	if modified, err := cacheModified(filename); err != nil {
		return err
	} else if modified {
		log.Println("Reading config file:", filename)

		// Open the file
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(config); err != nil {
			return err
		}

		config.ClosureLibrary = fixPath(config.ClosureLibrary)
		config.ClosureCompiler = fixPath(config.ClosureCompiler)
		config.ClosureTemplates = fixPath(config.ClosureTemplates)
		config.ClosureStylesheets = fixPath(config.ClosureStylesheets)
		config.Inherits = fixInheritsPath(filename, config.Inherits)
	}

	if config.Inherits != "" {
		if err := ReadFromFile(config.Inherits); err != nil {
			return err
		}
	}

	applyConf(config)

	return nil
}

func Validate() error {
	c := conf

	if c.Id == "" {
		return fmt.Errorf("the id of the app is required")
	}

	if c.RootJs == "" {
		return fmt.Errorf("the js root folder is required")
	}

	if c.Build == "" {
		return fmt.Errorf("the build folder is required")
	}

	if c.ClosureLibrary == "" || c.ClosureCompiler == "" || c.ClosureTemplates == "" ||
		c.ClosureStylesheets == "" {
		return fmt.Errorf("all the closure paths are required")
	}

	if c.Mode != "SIMPLE" && c.Mode != "ADVANCED" && c.Mode != "WHITESPACE" && c.Mode != "RAW" {
		return fmt.Errorf("illegal compilation mode: %s", c.Mode)
	}

	if c.Level != "QUIET" && c.Level != "DEFAULT" && c.Level != "VERBOSE" {
		return fmt.Errorf("illegal warning level: %s", c.Level)
	}

	if len(c.Inputs) == 0 {
		return fmt.Errorf("no inputs file provided")
	}

	if c.RenameCss != "true" && c.RenameCss != "false" && c.RootGss != "" {
		return fmt.Errorf("no renaming policy provided for gss")
	}

	for e, t := range c.Checks {
		checks := map[string]bool{
			"checkRegExp":            true,
			"checkTypes":             true,
			"checkVars":              true,
			"deprecated":             true,
			"fileoverviewTags":       true,
			"internetExplorerChecks": true,
			"invalidCasts":           true,
			"missingProperties":      true,
			"nonStandardJsDocs":      true,
			"typeInvalidation":       true,
			"undefinedVars":          true,
			"unknownDefines":         true,
			"uselessCode":            true,
		}
		if _, ok := checks[e]; !ok {
			return fmt.Errorf("illegal checj: %s", e)
		}

		if t != "WARNING" && t != "ERROR" && t != "OFF" {
			return fmt.Errorf("illegal value for the check %s: %s", e, t)
		}
	}

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

	if len(config.NonStandardCssFuncs) > 0 {
		conf.NonStandardCssFuncs = config.NonStandardCssFuncs
	}

	if config.RenameCss != "" {
		conf.RenameCss = config.RenameCss
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
