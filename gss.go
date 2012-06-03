package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

var cssTimesCache = map[string]time.Time{}

func CompileCssHandler(r *Request) error {
	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	// Compile the .gss files
	gss, err := ScanGss(conf.Root)
	if err != nil {
		return InternalErr(err, "cannot scan the root directory")
	}

	if err := CompileGss(r, gss); err != nil {
		return err
	}

	f, err := os.Open(path.Join(conf.Build, "compiled.css"))
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(r.W, f)

	return nil
}

// Scan a directory searching for .gss files
func ScanGss(filepath string) ([]string, error) {
	gss := []string{}

	// Get the list of entries
	ls, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil, err
	}

	for _, entry := range ls {
		fullpath := path.Join(filepath, entry.Name())

		if entry.IsDir() {
			if IsValidDir(entry.Name()) {
				// Scan recursively the directories
				f, err := ScanGss(fullpath)
				if err != nil {
					return nil, err
				}
				gss = append(gss, f...)
			}
		} else if path.Ext(entry.Name()) == ".gss" {
			// Add the templates to the list
			gss = append(gss, fullpath)
		}
	}

	return gss, nil
}

// Compile the stylesheets if they have been modified
func CompileGss(r *Request, gss []string) error {
	compiler := path.Join(conf.ClosureStylesheets, "build", "closure-stylesheets.jar")
	out := path.Join(conf.Build, "compiled.css")

	// Check if the cached version is still ok
	modified := false
	for _, filepath := range gss {
		info, err := os.Lstat(filepath)
		if err != nil {
			return err
		}

		t, ok := timesCache[filepath]
		if !ok || t != info.ModTime() {
			timesCache[filepath] = info.ModTime()
			modified = true
		}
	}

	if !modified {
		return nil
	}

	log.Println("Compiling gss")

	// Compile the template
	cmd := exec.Command("java", "-jar", compiler, "--output-file", out,
		"--output-renaming-map-format", "CLOSURE_COMPILED", "--rename", "CLOSURE",
		"--output-renaming-map", path.Join(conf.Build, "renaming-map.js"))

	for _, f := range gss {
		cmd.Args = append(cmd.Args, f)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return err
	}

	return nil
}
