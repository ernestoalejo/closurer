package config

import (
	"encoding/xml"
	"fmt"
	"os"
)

type OutputNode struct {
	Js  string `xml:"js,attr"`
	Css string `xml:"css,attr"`
}

type JsNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`

	Checks  ChecksNode     `xml:"checks"`
	Targets []JsTargetNode `xml:"target"`
	Inputs  []InputNode    `xml:"input"`
}

type ChecksNode struct {
	Errors   []CheckNode `xml:"error"`
	Warnings []CheckNode `xml:"warning"`
	Offs     []CheckNode `xml:"off"`
}

type CheckNode struct {
	Name string `xml:"name,attr"`
}

type JsTargetNode struct {
	Name  string `xml:"name,attr"`
	Mode  string `xml:"mode,attr"`
	Level string `xml:"level,attr"`

	Defines []DefineNode `xml:"define"`
}

type DefineNode struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type InputNode struct {
	File string `xml:"file,attr"`
}

type GssNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`

	Targets []GssTargetNode `xml:"target"`
}

type GssTargetNode struct {
	Name   string `xml:"name,attr"`
	Rename string `xml:"rename,attr"`

	Defines []DefineNode `xml:"define"`
}

type SoyNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`
}

type LibraryNode struct {
	Root string `xml:"root,attr"`
}

type Config struct {
	Build string `xml:"build,attr"`

	Output  OutputNode  `xml:"output"`
	Js      JsNode      `xml:"js"`
	Gss     GssNode     `xml:"gss"`
	Soy     SoyNode     `xml:"soy"`
	Library LibraryNode `xml:"library"`
}

// ==================================================================

func Load(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open the config file: %s", err)
	}
	defer f.Close()

	conf := new(Config)
	if err := xml.NewDecoder(f).Decode(&conf); err != nil {
		return nil, fmt.Errorf("cannot decode the config: %s", err)
	}

	return conf, nil
}

func (c *Config) Validate() error {
	if c.Js.Root == "" {
		return fmt.Errorf("The JS root folder is required")
	}
	if c.Build == "" {
		return fmt.Errorf("The build folder is required")
	}
	if c.Library.Root == "" {
		return fmt.Errorf("The Closure Library path is required")
	}
	if c.Js.Compiler == "" {
		return fmt.Errorf("The Closure Compiler path is required")
	}
	if c.Gss.Root != "" && c.Gss.Compiler == "" {
		return fmt.Errorf("The Closure Stylesheets path is required")
	}
	if c.Soy.Root != "" && c.Soy.Compiler == "" {
		return fmt.Errorf("The Closure Templates path is required")
	}

	for _, t := range c.Js.Targets {
		modes := map[string]bool{
			"SIMPLE":     true,
			"ADVANCED":   true,
			"WHITESPACE": true,
			"RAW":        true,
		}
		if _, ok := modes[t.Mode]; !ok {
			return fmt.Errorf("Illegal compilation mode in target %s: %s", t.Name, t.Mode)
		}

		levels := map[string]bool{
			"QUIET":   true,
			"DEFAULT": true,
			"VERBOSE": true,
		}
		if _, ok := levels[t.Level]; !ok {
			return fmt.Errorf("Illegal warning level in target %s: %s", t.Name, t.Level)
		}
	}

	if len(c.Js.Inputs) == 0 {
		return fmt.Errorf("Input files required in target")
	}

	for _, t := range c.Gss.Targets {
		if t.Rename != "true" && t.Rename != "false" && t.Rename != "" {
			return fmt.Errorf("Illegal renaming policy value")
		}
	}

	validChecks(c.Js.Checks.Errors)
	validChecks(c.Js.Checks.Warnings)
	validChecks(c.Js.Checks.Offs)

	return nil
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
			return fmt.Errorf("Illegal check: %s", check.Name)
		}
	}

	return nil
}
