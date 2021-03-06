package config

import (
	"encoding/xml"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ernestokarim/closurer/app"
)

var (
	globalConf       *Config
	lastModification time.Time
)

func Load() error {
	if globalConf != nil && !NoCache {
		info, err := os.Lstat(ConfPath)
		if err != nil {
			return app.Error(err)
		}

		if info.ModTime() == lastModification {
			return nil
		}
	}

	f, err := os.Open(ConfPath)
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	conf := new(Config)
	if err := xml.NewDecoder(f).Decode(&conf); err != nil {
		return app.Error(err)
	}

	// Assign it before validating, because we need it to
	// inherit targets.
	globalConf = conf

	if err := conf.validate(); err != nil {
		return err
	}

	info, err := os.Lstat(ConfPath)
	if err != nil {
		return app.Error(err)
	}
	lastModification = info.ModTime()

	return nil
}

func Current() *Config {
	return globalConf
}

func (c *Config) validate() error {
	// Library & compiler paths
	if c.Js != nil {
		if c.Js.Root == "" {
			return app.Errorf("The JS root folder is required")
		}

		if c.Js.Formatting != "" && c.Js.Formatting != "PRETTY_PRINT" {
			return app.Errorf("formatting mode not allowed: %s", c.Js.Formatting)
		}

		if c.Js.SideEffects != "" && c.Js.SideEffects != "true" {
			return app.Errorf("boolean value not allowed: %s", c.Js.SideEffects)
		}

		if c.Js.Language != "" {
			modes := map[string]bool{
				"ECMASCRIPT3":        true,
				"ECMASCRIPT5":        true,
				"ECMASCRIPT5_STRICT": true,
			}
			if _, ok := modes[c.Js.Language]; !ok {
				return app.Errorf("language mode not allowed: %s", c.Js.Language)
			}
		}

		// JS targets and inheritation
		if len(c.Js.Targets) == 0 {
			return app.Errorf("No target provided for JS code")
		}
		for _, t := range c.Js.Targets {
			if err := t.ApplyInherits(); err != nil {
				return err
			}
		}

		// Check compilation mode and warnings level
		for _, t := range c.Js.Targets {
			modes := map[string]bool{
				"SIMPLE":     true,
				"ADVANCED":   true,
				"WHITESPACE": true,
				"RAW":        true,
			}
			if _, ok := modes[t.Mode]; !ok {
				return app.Errorf("Illegal compilation mode in target %s: %s", t.Name, t.Mode)
			}

			levels := map[string]bool{
				"QUIET":   true,
				"DEFAULT": true,
				"VERBOSE": true,
			}
			if _, ok := levels[t.Level]; !ok {
				return app.Errorf("Illegal warning level in target %s: %s", t.Name, t.Level)
			}
		}

		// Check that the command line target is in the config file
		found := false
		for _, name := range TargetList() {
			for _, t := range c.Js.Targets {
				if t.Name == name {
					found = true
					break
				}
			}
			if !found {
				return app.Errorf("Target %s not found in the config file", name)
			}
		}

		// Validate the compilation checks
		if c.Js.Checks != nil {
			validChecks(c.Js.Checks.Errors)
			validChecks(c.Js.Checks.Warnings)
			validChecks(c.Js.Checks.Offs)
		}

		// Check the prepend files
		if c.Js.Prepends != nil {
			for _, prepend := range c.Js.Prepends {
				if prepend.File == "" {
					return app.Errorf("prepend file empty")
				}
			}
		}
	}

	if c.Build == "" {
		return app.Errorf("The build folder is required")
	}
	if c.Library != nil && c.Library.Root == "" {
		return app.Errorf("The Closure Library path is required")
	}
	if c.Js != nil && c.Js.Compiler == "" {
		return app.Errorf("The Closure Compiler path is required")
	}

	if c.Gss != nil {
		// GSS compiler
		if c.Gss.Compiler == "" {
			return app.Errorf("The Closure Stylesheets path is required")
		}

		// GSS targets
		if len(c.Gss.Targets) == 0 {
			return app.Errorf("No target provided for GSS code")
		}

		// At least one input file should be provided
		if len(c.Gss.Inputs) == 0 {
			return app.Errorf("No inputs provided for GSS code")
		}

		// Compare JS targets and GSS targets
		if c.Js != nil {
			if len(c.Js.Targets) != len(c.Gss.Targets) {
				return app.Errorf("Different number of targets provided for GSS & JS")
			}
			for i, tjs := range c.Js.Targets {
				tgss := c.Gss.Targets[i]
				if tjs.Name != tgss.Name {
					return app.Errorf("Targets with different name or order: %s != %s",
						tjs.Name, tgss.Name)
				}

				// Rename property of the GSS target
				if tgss.Rename != "true" && tgss.Rename != "false" && tgss.Rename != "" {
					return app.Errorf("Illegal renaming policy value")
				}

				// Apply the inherits option
				if err := tgss.ApplyInherits(); err != nil {
					return err
				}

				// Check that the GSS defines don't have a value
				for _, d := range tgss.Defines {
					if d.Value != "" {
						return app.Errorf("Define values in GSS should be empty")
					}
				}
			}
		}
	}

	// Soy compiler
	if c.Soy != nil && c.Soy.Root != "" && c.Soy.Compiler == "" {
		return app.Errorf("The Closure Templates path is required")
	}

	// Current targets in build mode
	if c.Js != nil && c.Gss != nil {
		for _, t := range TargetList() {
			SelectTarget(t)

			tjs := c.Js.CurTarget()
			tgss := c.Gss.CurTarget()

			if tjs == nil || tgss == nil {
				return app.Errorf("Target not found in the config: %s", t)
			}

			if Build && IsTarget(tjs.Name) {
				if tjs.Output == "" {
					return app.Errorf("Target to build JS without an output file: %s",
						tjs.Name)
				}
				if tgss != nil && tgss.Output == "" {
					return app.Errorf("Target to build GSS without an output file: %s",
						tjs.Name)
				}
			}
		}
	}

	// Fix the compilers paths
	if c.Js != nil {
		c.Js.Compiler = fixPath(c.Js.Compiler)
	}
	if c.Gss != nil {
		c.Gss.Compiler = fixPath(c.Gss.Compiler)
	}
	if c.Soy != nil {
		c.Soy.Compiler = fixPath(c.Soy.Compiler)
	}
	if c.Library != nil {
		c.Library.Root = fixPath(c.Library.Root)
	}

	return nil
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
		log.Fatal("Found ~ in a path, but USER nor USERNAME are exported in the env")
	}

	return strings.Replace(p, "~", "/home/"+user, -1)
}

func validChecks(lst []*CheckNode) error {
	for _, check := range lst {
		checks := map[string]bool{
			"ambiguousFunctionDecl":  true,
			"checkRegExp":            true,
			"checkTypes":             true,
			"checkVars":              true,
			"constantProperty":       true,
			"deprecated":             true,
			"fileoverviewTags":       true,
			"internetExplorerChecks": true,
			"invalidCasts":           true,
			"missingProperties":      true,
			"nonStandardJsDocs":      true,
			"strictModuleDepCheck":   true,
			"typeInvalidation":       true,
			"undefinedNames":         true,
			"undefinedVars":          true,
			"unknownDefines":         true,
			"uselessCode":            true,
			"globalThis":             true,
			"duplicateMessage":       true,
		}
		if _, ok := checks[check.Name]; !ok {
			return app.Errorf("Illegal check: %s", check.Name)
		}
	}

	return nil
}
