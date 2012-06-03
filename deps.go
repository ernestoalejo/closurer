package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"time"
)

var (
	provideRe  = regexp.MustCompile(`^\s*goog\.provide\(\s*[\'"](.+)[\'"]\s*\)`)
	requiresRe = regexp.MustCompile(`^\s*goog\.require\(\s*[\'"](.+)[\'"]\s*\)`)
	base       = "var goog = goog || {}; // Identifies this file as the Closure base."
)

var sourcesCache = map[string]*Source{}

// Saves the list of goog.provide() and goog.require() calls
// for each JS source.
type Source struct {
	provides []string
	requires []string
	base     bool
	modified time.Time
	filename string
}

// Creates a new source
func NewSource(filename string) (*Source, bool, error) {
	// Get the info of the file
	info, err := os.Lstat(filename)
	if err != nil {
		return nil, false, fmt.Errorf("cannot stat file info: %s: %s", filename, err)
	}

	// If it hasn't been modified, return in directly
	src, ok := sourcesCache[filename]
	if ok {
		if info.ModTime() == src.modified {
			return src, true, nil
		}
	}

	src = &Source{
		provides: []string{},
		requires: []string{},
		base:     false,
		filename: filename,
	}

	// Open the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, false, err
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
		matchs := provideRe.FindSubmatch(line)
		if matchs != nil {
			src.provides = append(src.provides, string(matchs[1]))
			continue
		}

		// Find the goog.require() calls
		matchs = requiresRe.FindSubmatch(line)
		if matchs != nil {
			src.requires = append(src.requires, string(matchs[1]))
			continue
		}

		// Recognize the base file
		if string(line) == base {
			src.base = true
		}
	}

	// Validates the base file
	if src.base {
		if len(src.provides) > 0 || len(src.requires) > 0 {
			return nil, false,
				fmt.Errorf("base files should not provide or require namespaces: %s", filename)
		}
		src.provides = append(src.provides, "goog")
	}

	// Save the file info in cache
	src.modified = info.ModTime()
	sourcesCache[filename] = src

	return src, false, nil
}

// Store the info of a dependencies tree
type DepsTree struct {
	sources     map[string]*Source
	provides    map[string]*Source
	base        *Source
	mustCompile bool
}

// Adds a new JS source file to the tree
func (tree *DepsTree) AddSource(filename string) error {
	// Build the source
	src, cached, err := NewSource(filename)
	if err != nil {
		return err
	}

	if src.base {
		tree.base = src
	}

	// Scan all the previous sources searching for repeated
	// namespaces
	for k, source := range tree.sources {
		for _, provide := range source.provides {
			if In(src.provides, provide) {
				return fmt.Errorf("multiple provide %s: %s and %s", provide, k, filename)
			}
		}
	}

	// Add all the provides to the list
	for _, provide := range src.provides {
		tree.provides[provide] = src
	}

	tree.sources[filename] = src
	tree.mustCompile = tree.mustCompile || !cached

	return nil
}

// Check if all required namespaces are provided by the 
// scanned files
func (tree *DepsTree) Check() error {
	for k, source := range tree.sources {
		for _, require := range source.requires {
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

	return src.provides, nil
}

// Struct to store the info of a dependencies tree traversal
type TraversalInfo struct {
	deps      []*Source
	traversal []string
}

// Returns the list of files (in order) that must be compiled to finally
// obtain all namespaces, including the base one.
func (tree *DepsTree) GetDependencies(namespaces []string) ([]*Source, error) {
	depsList := []*Source{tree.base}

	for _, ns := range namespaces {
		// Prepare the info
		info := &TraversalInfo{
			deps:      []*Source{},
			traversal: []string{},
		}

		// Resolve all the needed dependencies
		if err := tree.ResolveDependencies(ns, info); err != nil {
			return nil, err
		}

		// Add it to the list if they're not there yet
		for _, k := range info.deps {
			if !InSource(depsList, k) {
				depsList = append(depsList, k)
			}
		}
	}

	return depsList, nil
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
		for _, require := range src.requires {
			tree.ResolveDependencies(require, info)
		}

		// Add ourselves to the list of files
		info.deps = append(info.deps, src)

		// Remove the namespace from the traversal
		info.traversal = info.traversal[:len(info.traversal)-1]
	}

	return nil
}

// Build a dependency tree that allows the client to know the order of
// compilation
func BuildDepsTree(r *Request) (*DepsTree, error) {
	// Roots directories
	roots := append([]string{
		conf.Root,
		conf.ClosureLibrary,
		path.Join(conf.ClosureTemplates, "javascript"),
		path.Join(conf.Build, "templates"),
	}, conf.Paths...)

	// Build the deps tree scanning each root directory recursively
	depstree := &DepsTree{
		sources:  map[string]*Source{},
		provides: map[string]*Source{},
	}
	for _, root := range roots {
		if err := ScanSources(depstree, root); err != nil {
			return nil, err
		}
	}

	// Check the integrity of the tree
	if err := depstree.Check(); err != nil {
		return nil, err
	}

	return depstree, nil
}

func ScanSources(depstree *DepsTree, filepath string) error {
	// Read the directory contents
	ls, err := ioutil.ReadDir(filepath)
	if err != nil {
		return err
	}

	// Scan them
	for _, entry := range ls {
		fullpath := path.Join(filepath, entry.Name())

		if entry.IsDir() {
			if IsValidDir(entry.Name()) {
				log.Println("valid: ", entry.Name())
				// Scan directories recursively
				if err := ScanSources(depstree, fullpath); err != nil {
					return err
				}
			}
		} else if path.Ext(entry.Name()) == ".js" {
			// Add sources to the list
			if err := depstree.AddSource(fullpath); err != nil {
				return err
			}
		}
	}

	return nil
}
