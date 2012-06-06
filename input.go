package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SourcesList []*Source

func (lst SourcesList) Len() int {
	return len(lst)
}

func (lst SourcesList) Less(i, j int) bool {
	return lst[i].filename < lst[j].filename
}

func (lst SourcesList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}

func InputHandler(r *Request) error {
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

	// Calculate the list of files to compile
	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return err
	}

	log.Println("Done generating deps.js! Elapsed:", time.Since(start))

	// Sort the list of dependencies (the order doesn't mind)
	sorted_deps := SourcesList(deps)
	sort.Sort(sorted_deps)

	for _, src := range sorted_deps {
		// Accumulates the provides
		provides := ""
		for _, provide := range src.provides {
			provides += "'" + provide + "' "
		}

		// Accumulates the requires
		requires := ""
		for _, require := range src.requires {
			requires += "'" + require + "' "
		}

		// Search the base path to the file, and put the path
		// relative to it
		var n string
		for _, p := range paths {
			n, err = filepath.Rel(p, src.filename)
			if err == nil && !strings.Contains(n, "..") {
				break
			}
		}
		if n == "" {
			return fmt.Errorf("cannot generate the relative filename for %s", src.filename)
		}

		// Write the line to the output of the deps.js file request
		fmt.Fprintf(r.W, "goog.addDependency('%s', [%s], [%s]);\n", n, provides, requires)
	}

	return nil
}
