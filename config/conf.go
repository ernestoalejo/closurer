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

	if err := conf.validate(); err != nil {
		return err
	}

	globalConf = conf

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
	if c.Js.Root == "" {
		return app.Errorf("The JS root folder is required")
	}
	if c.Build == "" {
		return app.Errorf("The build folder is required")
	}
	if c.Library.Root == "" {
		return app.Errorf("The Closure Library path is required")
	}
	if c.Js.Compiler == "" {
		return app.Errorf("The Closure Compiler path is required")
	}
	if len(c.Js.Targets) == 0 {
		return app.Errorf("No target provided for JS code")
	}
	if c.Gss.Root != "" {
		if c.Gss.Compiler == "" {
			return app.Errorf("The Closure Stylesheets path is required")
		}
		if len(c.Gss.Targets) == 0 {
			return app.Errorf("No target provided for GSS code")
		}
		if len(c.Js.Targets) != len(c.Gss.Targets) {
			return app.Errorf("Different number of targets provided for GSS & JS")
		}

		for i, tjs := range c.Js.Targets {
			tgss := c.Gss.Targets[i]
			if tjs.Name != tgss.Name {
				return app.Errorf("Targets with different name or order: %s != %s",
					tjs.Name, tgss.Name)
			}
		}
	}
	if c.Soy.Root != "" && c.Soy.Compiler == "" {
		return app.Errorf("The Closure Templates path is required")
	}

	tjs := c.Js.CurTarget()
	tgss := c.Gss.CurTarget()
	if Build && tjs.Name == Target {
		if tjs.Output == "" {
			return app.Errorf("Target to build JS without an output file: %s",
				tjs.Name)
		}
		if tgss != nil && tgss.Output == "" {
			return app.Errorf("Target to build GSS without an output file: %s",
				tjs.Name)
		}
	}

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

	if len(c.Js.Inputs) == 0 {
		return app.Errorf("Input files required in target")
	}

	for _, t := range c.Gss.Targets {
		if t.Rename != "true" && t.Rename != "false" && t.Rename != "" {
			return app.Errorf("Illegal renaming policy value")
		}
	}

	found := false
	for _, t := range c.Js.Targets {
		if t.Name == Target {
			found = true
			break
		}
	}

	if !found {
		return app.Errorf("Target %s not found in the config file", Target)
	}

	validChecks(c.Js.Checks.Errors)
	validChecks(c.Js.Checks.Warnings)
	validChecks(c.Js.Checks.Offs)

	c.Gss.Compiler = fixPath(c.Gss.Compiler)
	c.Js.Compiler = fixPath(c.Js.Compiler)
	c.Soy.Compiler = fixPath(c.Soy.Compiler)
	c.Library.Root = fixPath(c.Library.Root)

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

func validChecks(lst []CheckNode) error {
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
