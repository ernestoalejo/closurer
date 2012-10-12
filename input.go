package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
)

func Input(r *app.Request) error {
	// app.Requested filename
	name := r.Req.URL.Path[7:]

	// Execute the pre-compile actions
	if err := PreCompileActions(); err != nil {
		return err
	}

	// Re-calculate deps and compile templates if needed
	if name == "deps.js" {
		return GenerateDeps(r)
	}

	// Otherwise serve the file if it can be found
	paths := BaseJSPaths()
	for _, p := range paths {
		f, err := os.Open(path.Join(p, name))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot open the file: %s", err)
		} else if err == nil {
			defer f.Close()

			r.W.Header().Set("Content-Type", "text/javascript")
			io.Copy(r.W, f)

			// Execute the post-compile actions
			if err := PostCompileActions(); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("file not found: %s", name)
}

func GenerateDeps(r *app.Request) error {
	// Execute the pre-compile actions
	if err := PreCompileActions(); err != nil {
		return err
	}

	// Compile all the modified templates
	if err := CompileSoy(); err != nil {
		return err
	}

	start := time.Now()
	log.Println("Building dependency tree...")

	// Build the dependency tree between the JS files
	depstree, err := NewDepsTree("input")
	if err != nil {
		return err
	}

	conf := config.Current()

	// Calculate all the input namespaces
	namespaces := []string{}
	for _, input := range conf.Inputs {
		ns, err := depstree.GetProvides(input)
		if err != nil {
			return err
		}
		namespaces = append(namespaces, ns...)
	}

	// Add some special namespaces for easier testing
	namespaces = append(namespaces, "goog.userAgent.product",
		"goog.testing.MultiTestRunner")

	// Calculate the list of dependencies
	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return err
	}

	log.Println("Done generating deps.js! Elapsed:", time.Since(start))

	// Execute the post-compile actions
	if err := PostCompileActions(); err != nil {
		return err
	}

	// Output the list correctly formatted
	r.W.Header().Set("Content-Type", "text/javascript")
	if err := WriteDeps(r.W, deps); err != nil {
		return err
	}

	return nil
}
