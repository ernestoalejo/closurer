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
	// Reload the confs if they've changed
	if err := ReadConf(); err != nil {
		return err
	}

	// Compile the code
	if err := CompileJs(r.W); err != nil {
		return err
	}

	f, err := os.Open(path.Join(conf.Build, "compiled.js"))
	if err != nil {
		return fmt.Errorf("cannot read the compiled javascript: %s", err)
	}
	defer f.Close()

	r.W.Header().Set("Content-Type", "text/javascript")
	io.Copy(r.W, f)

	return nil
}

func CompileJs(w io.Writer) error {
	start := time.Now()

	// Compile the .gss files
	if err := CompileCss(w); err != nil {
		return err
	}

	// Compile the .soy files
	if err := CompileSoy(w); err != nil {
		return err
	}

	// Build the dependency tree between the JS files
	depstree, err := BuildDepsTree()
	if err != nil {
		return err
	}

	mustCompile := false

	out := path.Join(conf.Build, "compiled.js")
	if *build {
		out = *jsOutput
		mustCompile = true
	}

	if !mustCompile {
		if _, err = os.Lstat(out); err != nil {
			if os.IsNotExist(err) {
				mustCompile = true
			} else {
				return err
			}
		}
	}

	if mustCompile || depstree.mustCompile {
		// Calculate all the input namespaces
		namespaces := []string{}
		for _, input := range conf.Inputs {
			if strings.Contains(input, "_test") {
				continue
			}

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

		f, err := os.Create(path.Join(conf.Build, "deps.js"))
		if err != nil {
			return fmt.Errorf("cannot create deps file: %s", err)
		}
		defer f.Close()

		// Base paths, all routes to a JS must start from these
		paths := []string{
			path.Join(conf.ClosureLibrary, "closure", "goog"),
			conf.RootJs,
			path.Join(conf.Build, "templates"),
			conf.RootSoy,
			path.Join(conf.ClosureTemplates, "javascript"),
		}

		if err := WriteDeps(f, deps, paths); err != nil {
			return err
		}

		// Send them to the compiler
		if err := JsCompiler(w, deps); err != nil {
			return err
		}
	}

	log.Println("Done compiling! Elapsed:", time.Since(start))

	return nil
}

func JsCompiler(w io.Writer, deps []*Source) error {
	out := path.Join(conf.Build, "compiled.js")
	if *build {
		out = *jsOutput
	}

	// Prepare the call to the compiler
	args := []string{
		"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--js_output_file", out,
		"--js", path.Join(conf.Build, "renaming-map.js"),
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "deps.js"),
		"--js", path.Join(conf.Build, "deps.js"),
	}

	// Add the dependencies in order
	for _, dep := range deps {
		if !strings.Contains(dep.Filename, "_test.js") {
			args = append(args, "--js", dep.Filename)
		}
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

	// Add the externs
	for _, extern := range conf.Externs {
		args = append(args, "--externs", extern)
	}

	if *outputCmd {
		f, err := os.Create(path.Join(conf.Build, "cmd"))
		if err != nil {
			return fmt.Errorf("cannot create the output command file: %s", err)
		}
		fmt.Fprintln(f, args)
		f.Close()
	}

	log.Println("Compiling code to build/compiled.js")

	// Compile the code
	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		//log.Println("Output from compiler:\n", string(output))
		fmt.Fprintf(w, "%s\n", output)
		return fmt.Errorf("cannot compile the code: %s", err)
	}

	if len(output) > 0 {
		log.Println("Output from compiler:\n", string(output))
	}

	return nil
}
