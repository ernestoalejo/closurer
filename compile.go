package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/hooks"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/soy"
)

func Compile(r *app.Request) error {
	conf := config.Current()

	// Execute the pre-compile actions
	if err := hooks.PreCompile(); err != nil {
		return err
	}

	if conf.Mode == "RAW" {
		if err := RawOutput(r); err != nil {
			return err
		}
	} else {
		// Compile the code
		if err := CompileJs(r.W); err != nil {
			return err
		}
	}

	// Execute the post-compile actions
	if err := hooks.PostCompile(); err != nil {
		return err
	}

	if conf.Mode != "RAW" {
		// Copy the file to the output
		f, err := os.Open(path.Join(conf.Build, "compiled.js"))
		if err != nil {
			return app.Error(err)
		}
		defer f.Close()

		r.W.Header().Set("Content-Type", "text/javascript")
		io.Copy(r.W, f)
	}

	return nil
}

func CompileJs(w io.Writer) error {
	if err := gss.Compile(); err != nil {
		return err
	}

	if err := copyCssFile(); err != nil {
		return err
	}

	// Compile the .soy files
	if err := soy.Compile(); err != nil {
		return err
	}

	// Build the dependency tree between the JS files
	depstree, err := scan.NewDepsTree("compile")
	if err != nil {
		return err
	}

	// Whether we must recompile or the old file is correct
	mustCompile := false

	conf := config.Current()

	// Build the out path
	out := path.Join(conf.Build, "compiled.js")
	if config.Build {
		out = *jsOutput
		mustCompile = true
	}

	if !mustCompile {
		// Check if the cached file exists, to use it
		if _, err = os.Lstat(out); err != nil && os.IsNotExist(err) {
			mustCompile = true
		} else if err != nil {
			return err
		}
	}

	if mustCompile || depstree.MustCompile {
		// Calculate all the input namespaces
		namespaces := []string{}
		for _, input := range conf.Inputs {
			// Ignore _test files
			if strings.Contains(input, "_test") {
				continue
			}

			ns, err := depstree.GetProvides(input)
			if err != nil {
				return err
			}
			namespaces = append(namespaces, ns...)
		}

		// Calculate the list of files to compile
		deps, err := depstree.GetDependencies(namespaces)
		if err != nil {
			return err
		}

		// Create the deps.js file for our project
		f, err := os.Create(path.Join(conf.Build, "deps.js"))
		if err != nil {
			return app.Error(err)
		}
		defer f.Close()

		// Write the deps file
		if err := scan.WriteDeps(f, deps); err != nil {
			return err
		}

		// Compile the javascript
		if err := js.Compile(out, deps); err != nil {
			return err
		}
	}

	return nil
}

func copyCssFile() error {
	conf := config.Current()

	src, err := os.Open(filepath.Join(conf.Build, gss.CSS_NAME))
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(*cssOutput)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}
