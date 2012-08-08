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
	r.W.Header().Set("Content-Type", "text/css")

	// Output early if there's no GSS files
	if conf.RootGss == "" {
		fmt.Fprintln(r.W, "")
		return nil
	}

	// Execute the pre-compile actions
	if err := PreCompileActions(); err != nil {
		return err
	}

	// Compile the .gss files
	if err := CompileGss(); err != nil {
		return err
	}

	// Execute the post-compile actions
	if err := PostCompileActions(); err != nil {
		return err
	}

	// Copy the compile file to the output
	f, err := os.Open(path.Join(conf.Build, "compiled.css"))
	if err != nil {
		return fmt.Errorf("cannot read the compiled css: %s", err)
	}
	defer f.Close()

	io.Copy(r.W, f)

	return nil
}

// Compiles the .gss files
func CompileGss() error {
	// Output early if there's no GSS files.
	if conf.RootGss == "" {
		return nil
	}

	// Search the .gss files
	gss, err := Scan(conf.RootGss, ".gss")
	if err != nil {
		return err
	}

	// No results, no compiling
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

	if !modified && !*build {
		return nil
	}

	log.Println("Compiling gss")

	// Compute the output path
	out := path.Join(conf.Build, "compiled.css")
	if *build {
		out = *cssOutput
	}

	// Run the soy compiler
	cmd := exec.Command(
		"java",
		"-jar", path.Join(conf.ClosureStylesheets, "build", "closure-stylesheets.jar"),
		"--output-file", out,
		"--output-renaming-map-format", "CLOSURE_COMPILED",
		"--rename", "CLOSURE",
		"--output-renaming-map", path.Join(conf.Build, "renaming-map.js"))
	cmd.Args = append(cmd.Args, gss...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gss compiler error: %s\n%s", err, string(output))
	}

	return nil
}
