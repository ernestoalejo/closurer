package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
)

// Returns the compiled CSS.
func CompileGssHandler(r *Request) error {
	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	// Compile the .gss files
	if err := CompileGss(); err != nil {
		return err
	}

	// Copy the compile file to the output
	f, err := os.Open(path.Join(conf.Build, "compiled.css"))
	if err != nil {
		return fmt.Errorf("cannot read the compiled css: %s", err)
	}
	defer f.Close()

	r.W.Header().Set("Content-Type", "text/css")
	io.Copy(r.W, f)

	return nil
}

// Compiles the .gss files
func CompileGss() error {
	// Search the .gss files
	gss, err := Scan(conf.RootGss, ".gss")
	if err != nil {
		return err
	}

	// No result, no compiling
	if len(gss) == 0 {
		return nil
	}

	// Check if the cached version is still ok
	modified := false
	for _, filepath := range gss {
		if m, err := CacheModified(filepath); err != nil {
			return err
		} else if m {
			modified = true
			break
		}
	}

	if !modified {
		return nil
	}

	log.Println("Compiling gss")

	// Compute some paths
	compiler := path.Join(conf.ClosureStylesheets, "build", "closure-stylesheets.jar")

	out := path.Join(conf.Build, "compiled.css")
	if *build {
		out = *cssOutput
	}

	// Run the soy compiler
	cmd := exec.Command(
		"java", "-jar", compiler,
		"--output-file", out,
		"--output-renaming-map-format", "CLOSURE_COMPILED",
		"--rename", "CLOSURE",
		"--output-renaming-map", path.Join(conf.Build, "renaming-map.js"))
	cmd.Args = append(cmd.Args, gss...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gss compiler error:\n%s", output)
	}

	return nil
}
