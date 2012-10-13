package main

import (
	"io"
	"log"
	"os"
	"path"

	"github.com/gorilla/mux"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/hooks"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/soy"
)

func Input(r *app.Request) error {
	name := mux.Vars(r.Req)["name"]

	// Execute the pre-compile actions
	if err := hooks.PreCompile(); err != nil {
		return err
	}

	// Re-calculate deps and compile templates if needed
	if name == "deps.js" {
		return GenerateDeps(r)
	}

	// Otherwise serve the file if it can be found
	paths := scan.BaseJSPaths()
	for _, p := range paths {
		f, err := os.Open(path.Join(p, name))
		if err != nil && !os.IsNotExist(err) {
			return app.Error(err)
		} else if err == nil {
			defer f.Close()

			r.W.Header().Set("Content-Type", "text/javascript")
			io.Copy(r.W, f)

			// Execute the post-compile actions
			if err := hooks.PostCompile(); err != nil {
				return err
			}

			return nil
		}
	}

	return app.Errorf("file not found: %s", name)
}

func GenerateDeps(r *app.Request) error {
	if err := hooks.PreCompile(); err != nil {
		return err
	}

	if err := soy.Compile(); err != nil {
		return err
	}

	log.Println("Building dependency tree...")

	depstree, err := scan.NewDepsTree("input")
	if err != nil {
		return err
	}

	conf := config.Current()

	namespaces := []string{}
	for _, input := range conf.Inputs {
		ns, err := depstree.GetProvides(input)
		if err != nil {
			return err
		}
		namespaces = append(namespaces, ns...)
	}

	namespaces = append(namespaces, "goog.userAgent.product",
		"goog.testing.MultiTestRunner")

	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return err
	}

	log.Println("Done generating deps.js!")

	if err := hooks.PostCompile(); err != nil {
		return err
	}

	r.W.Header().Set("Content-Type", "text/javascript")
	if err := scan.WriteDeps(r.W, deps); err != nil {
		return err
	}

	return nil
}
