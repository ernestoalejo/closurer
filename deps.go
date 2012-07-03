package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	provideRe  = regexp.MustCompile(`^\s*goog\.provide\(\s*[\'"](.+)[\'"]\s*\)`)
	requiresRe = regexp.MustCompile(`^\s*goog\.require\(\s*[\'"](.+)[\'"]\s*\)`)
	//base       = "var goog = goog || {}; // Identifies this file as the Closure base."
)

// Represents a JS source
type Source struct {
	// List of namespaces this file provides.
	Provides []string

	// List of required namespaces for this file.
	Requires []string

	// Whether this is the base.js file of the Closure Library.
	Base bool

	// Name of the source file.
	Filename string
}

// Creates a new source. Returns the source, if it has been
// loaded from cache or not, and an error.
func NewSource(filename string, base string) (*Source, bool, error) {
	src := CachedSource(filename)

	// Return the file from cache if possible
	if modified, err := CacheModified(filename); err != nil {
		return nil, false, err
	} else if !modified {
		return src, true, nil
	}

	// Reset the source info
	src.Provides = []string{}
	src.Requires = []string{}
	src.Base = (filename == base)
	src.Filename = filename

	// Open the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, false, fmt.Errorf("cannot open the source file %s: %s", filename, err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		// Read it line by line
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, false, err
		}

		// Find the goog.provide() calls
		if strings.Contains(string(line), "goog.provide") {
			matchs := provideRe.FindSubmatch(line)
			if matchs != nil {
				src.Provides = append(src.Provides, string(matchs[1]))
				continue
			}
		}

		// Find the goog.require() calls
		if strings.Contains(string(line), "goog.require") {
			matchs := requiresRe.FindSubmatch(line)
			if matchs != nil {
				src.Requires = append(src.Requires, string(matchs[1]))
				continue
			}
		}
	}

	// Validates the base file
	if src.Base {
		if len(src.Provides) > 0 || len(src.Requires) > 0 {
			return nil, false,
				fmt.Errorf("base files should not provide or require namespaces: %s", filename)
		}
		src.Provides = append(src.Provides, "goog")
	}

	return src, false, nil
}

// Store the info of a dependencies tree
type DepsTree struct {
	sources     map[string]*Source
	provides    map[string]*Source
	base        *Source
	basePath    string
	mustCompile bool
}

// Build a dependency tree that allows the client to know the order of
// compilation
func NewDepsTree() (*DepsTree, error) {
	// Initialize the tree
	depstree := &DepsTree{
		sources:  map[string]*Source{},
		provides: map[string]*Source{},
		basePath: path.Join(conf.ClosureLibrary, "closure", "goog", "base.js"),
	}

	// Build the deps tree scanning each root directory recursively
	roots := BaseJSPaths(false)
	for _, root := range roots {
		// Scan the sources
		src, err := Scan(root, ".js")
		if err != nil {
			return nil, err
		}

		// Add them to the tree
		for _, s := range src {
			if err := depstree.AddSource(s); err != nil {
				return nil, err
			}
		}
	}

	// Check the integrity of the tree
	if err := depstree.Check(); err != nil {
		return nil, err
	}

	return depstree, nil
}

// Adds a new JS source file to the tree
func (tree *DepsTree) AddSource(filename string) error {
	// Build the source
	src, cached, err := NewSource(filename, tree.basePath)
	if err != nil {
		return err
	}

	// If it's the base file, save it
	if src.Base {
		tree.base = src
	}

	// Scan all the previous sources searching for repeated
	// namespaces. We ignore closure library files because they're
	// supposed to be correct and tested by other methods
	if !strings.HasPrefix(filename, conf.ClosureLibrary) {
		for k, source := range tree.sources {
			for _, provide := range source.Provides {
				if In(src.Provides, provide) {
					return fmt.Errorf("multiple provide %s: %s and %s", provide, k, filename)
				}
			}
		}
	}

	// Add all the provides to the map
	for _, provide := range src.Provides {
		tree.provides[provide] = src
	}

	// Save the source
	tree.sources[filename] = src

	// Update the mustCompile flag
	tree.mustCompile = tree.mustCompile || !cached

	return nil
}

// Check if all required namespaces are provided by the 
// scanned files
func (tree *DepsTree) Check() error {
	for k, source := range tree.sources {
		for _, require := range source.Requires {
			_, ok := tree.provides[require]
			if !ok {
				return fmt.Errorf("namespace not found %s: %s", require, k)
			}
		}
	}

	return nil
}

// Returns the provides list of a source file, or an error if it hasn't been
// scanned previously into the tree
func (tree *DepsTree) GetProvides(filename string) ([]string, error) {
	src, ok := tree.sources[filename]
	if !ok {
		return nil, fmt.Errorf("input not present in the sources: %s", filename)
	}

	return src.Provides, nil
}

// Struct to store the info of a dependencies tree traversal
type TraversalInfo struct {
	deps      []*Source
	traversal []string
}

// Returns the list of files (in order) that must be compiled to finally
// obtain all namespaces, including the base one.
func (tree *DepsTree) GetDependencies(namespaces []string) ([]*Source, error) {
	// Prepare the info
	info := &TraversalInfo{
		deps:      []*Source{},
		traversal: []string{},
	}

	for _, ns := range namespaces {
		// Resolve all the needed dependencies
		if err := tree.ResolveDependencies(ns, info); err != nil {
			return nil, err
		}
	}

	return info.deps, nil
}

// Adds to the traversal info the list of dependencies recursively.
func (tree *DepsTree) ResolveDependencies(ns string, info *TraversalInfo) error {
	// Check that the namespace is correct
	src, ok := tree.provides[ns]
	if !ok {
		return fmt.Errorf("namespace not found: %s", ns)
	}

	// Detects circular deps
	if In(info.traversal, ns) {
		info.traversal = append(info.traversal, ns)
		return fmt.Errorf("circular dependency detected: %v", info.traversal)
	}

	// Memoize results, don't recalculate old depencies
	if !InSource(info.deps, src) {
		// Add a new namespace to the traversal
		info.traversal = append(info.traversal, ns)

		// Compile first all dependencies
		for _, require := range src.Requires {
			tree.ResolveDependencies(require, info)
		}

		// Add ourselves to the list of files
		info.deps = append(info.deps, src)

		// Remove the namespace from the traversal
		info.traversal = info.traversal[:len(info.traversal)-1]
	}

	return nil
}

func WriteDeps(f io.Writer, deps []*Source) error {
	paths := BaseJSPaths(true)
	for _, src := range deps {
		// Accumulates the provides & requires of the source
		provides := "'" + strings.Join(src.Provides, "', '") + "'"
		requires := "'" + strings.Join(src.Requires, "', '") + "'"

		// Search the base path to the file, and put the path
		// relative to it
		var n string
		for _, p := range paths {
			var err error
			n, err = filepath.Rel(p, src.Filename)
			if err == nil && !strings.Contains(n, "..") {
				break
			}
		}
		if n == "" {
			return fmt.Errorf("cannot generate the relative filename for %s", src.Filename)
		}

		// Write the line to the output of the deps.js file request
		fmt.Fprintf(f, "goog.addDependency('%s', [%s], [%s]);\n", n, provides, requires)
	}

	return nil
}

// Base paths, all routes to a JS must start from one
// of these ones.
// The order is important, the paths will be scanned as
// they've been written.
// The param library adds the subdirectoy prefix to the Closure
// Library path if it's true.
func BaseJSPaths(library bool) []string {
	closureLibrary := path.Join(conf.ClosureLibrary)
	if library {
		closureLibrary = path.Join(closureLibrary, "closure", "goog")
	}

	return []string{
		closureLibrary,
		conf.RootJs,
		path.Join(conf.Build, "templates"),
		path.Join(conf.ClosureTemplates, "javascript"),
	}
}
