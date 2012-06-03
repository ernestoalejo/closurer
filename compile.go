package main

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"
)

var cachedOutput string

func CompileHandler(r *Request) error {
	start := time.Now()

	// Compile the .soy files
	if err := CompileTemplates(r); err != nil {
		return err
	}

	// Build the dependency tree between the JS files
	depstree, err := BuildDepsTree(r)
	if err != nil {
		return err
	}

	if depstree.mustCompile {
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

		// Send them to the compiler
		if err := CompileCode(r, deps); err != nil {
			return err
		}
	}

	log.Println("Done compiling! Elapsed:", time.Since(start))

	fmt.Fprintf(r.W, "%s\n", cachedOutput)

	return nil
}

func CompileCode(r *Request, deps []*Source) error {
	// Prepare the call to the compiler
	args := []string{"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar")}

	// Add the dependencies in order
	for _, dep := range deps {
		args = append(args, "--js", dep.filename)
	}

	// Add the defines
	for k, define := range conf.Define {
		if define != "true" && define != "false" {
			define = "\"" + define + "\""
		}
		args = append(args, "--define", k+"="+define)
	}

	// Add the compilation mode
	if conf.Mode == "ADVANCED" {
		args = append(args, "--compilation_level", "ADVANCED_OPTIMIZATIONS")
	} else if conf.Mode == "SIMPLE" {
		args = append(args, "--compilation_level", "SIMPLE_OPTIMIZATIONS")
	} else if conf.Mode == "WHITESPACE" {
		args = append(args, "--compilation_level", "WHITESPACE_ONLY")
	} else {
		return fmt.Errorf("compilation mode not recognized: %s", conf.Mode)
	}

	// Add the warning level
	if conf.Level == "QUIET" || conf.Level == "DEFAULT" || conf.Level == "VERBOSE" {
		args = append(args, "--warning_level", conf.Level)
	} else {
		return fmt.Errorf("warnings level not recognized: %s", conf.Level)
	}

	log.Println("Compiling code to build/compiled.js")

	// Compile the code
	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, "cannot compile the code")
	}

	// Cache the output for later re-use
	cachedOutput = string(output)

	return nil
}
