package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type SourcesList []*Source

func (lst SourcesList) Len() int {
	return len(lst)
}

func (lst SourcesList) Less(i, j int) bool {
	return lst[i].Filename < lst[j].Filename
}

func (lst SourcesList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}

func InputHandler(r *Request) error {
	r.W.Header().Set("Content-Type", "text/javascript")

	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	// Filename
	name := r.Req.URL.Path[7:]

	// Base paths, all routes to a JS must start from these
	paths := []string{
		path.Join(conf.ClosureLibrary, "closure", "goog"),
		conf.RootJs,
		path.Join(conf.Build, "templates"),
		conf.RootSoy,
		path.Join(conf.ClosureTemplates, "javascript"),
	}

	// Re-calculate deps and compile templates if needed
	if name == "deps.js" {
		return GenerateDeps(r, name, paths)
	}

	// Otherwise serve the file if it can be found
	for _, p := range paths {
		f, err := os.Open(path.Join(p, name))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot open the file: %s", err)
		} else if !os.IsNotExist(err) {
			defer f.Close()
			io.Copy(r.W, f)
			return nil
		}
	}

	return fmt.Errorf("file not found: %s", name)
}

func GenerateDeps(r *Request, name string, paths []string) error {
	// Compile all the modified templates
	if err := CompileSoy(r.W); err != nil {
		return err
	}

	start := time.Now()
	log.Println("Building dependency tree...")

	// Build the dependency tree between the JS files
	depstree, err := BuildDepsTree()
	if err != nil {
		return err
	}

	// Calculate all the input namespaces
	namespaces := []string{}
	for _, input := range conf.Inputs {
		ns, err := depstree.GetProvides(input)
		if err != nil {
			return err
		}
		namespaces = append(namespaces, ns...)
	}

	namespaces = append(namespaces, "goog.userAgent.product", "goog.testing.MultiTestRunner")

	// Calculate the list of files to compile
	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return err
	}

	log.Println("Done generating deps.js! Elapsed:", time.Since(start))

	if err := WriteDeps(r.W, deps, paths); err != nil {
		return err
	}

	return nil
}
