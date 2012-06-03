package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func CompileHandler(r *Request) error {
	start := time.Now()

	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

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

	f, err := os.Open(path.Join(conf.Build, "compiled.js"))
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(r.W, f)

	return nil
}

func CompileCode(r *Request, deps []*Source) error {
	// Prepare the call to the compiler
	args := []string{
		"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--js_output_file", path.Join(conf.Build, "compiled.js"),
	}

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

	// Add the checks
	for k, check := range conf.Checks {
		if check != "OFF" && check != "ERROR" && check != "WARNING" {
			return fmt.Errorf("unrecognized compiler check: %s", check)
		}
		args = append(args, "--jscomp_"+strings.ToLower(check), k)
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

	if *outputCmd {
		f, err := os.Create(path.Join(conf.Build, "cmd"))
		if err != nil {
			return err
		}
		fmt.Fprintln(f, args)
		f.Close()
	}

	log.Println("Compiling code to build/compiled.js")

	// Compile the code
	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Output from compiler:\n", string(output))
		fmt.Fprintf(r.W, "%s\n", output)
		return fmt.Errorf("cannot compile the code: %s", err)
	}

	if len(output) > 0 {
		log.Println("Output from compiler:\n", string(output))
	}

	return nil
}
