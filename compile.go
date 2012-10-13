package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/hooks"
)

func Compile(r *app.Request) error {
	conf := config.Current()

	// Execute the pre-compile actions
	if err := hooks.PreCompile(); err != nil {
		return err
	}

	if conf.Mode == "RAW" {
		if err := RawOutput(r); err != nil {
			return err
		}
	} else {
		// Compile the code
		if err := CompileJs(r.W); err != nil {
			return err
		}
	}

	// Execute the post-compile actions
	if err := hooks.PostCompile(); err != nil {
		return err
	}

	if conf.Mode != "RAW" {
		// Copy the file to the output
		f, err := os.Open(path.Join(conf.Build, "compiled.js"))
		if err != nil {
			return app.Error(err)
		}
		defer f.Close()

		r.W.Header().Set("Content-Type", "text/javascript")
		io.Copy(r.W, f)
	}

	return nil
}

func CompileJs(w io.Writer) error {
	start := time.Now()

	// Compile the .gss files
	if err := gss.Compile(); err != nil {
		return err
	}

	if err := copyCssFile(); err != nil {
		return err
	}

	// Compile the .soy files
	if err := CompileSoy(); err != nil {
		return err
	}

	// Build the dependency tree between the JS files
	depstree, err := NewDepsTree("compile")
	if err != nil {
		return err
	}

	// Whether we must recompile or the old file is correct
	mustCompile := false

	conf := config.Current()

	// Build the out path
	out := path.Join(conf.Build, "compiled.js")
	if config.Build {
		out = *jsOutput
		mustCompile = true
	}

	if !mustCompile {
		// Check if the cached file exists, to use it
		if _, err = os.Lstat(out); err != nil && os.IsNotExist(err) {
			mustCompile = true
		} else if err != nil {
			return err
		}
	}

	if mustCompile || depstree.mustCompile {
		// Calculate all the input namespaces
		namespaces := []string{}
		for _, input := range conf.Inputs {
			// Ignore _test files
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

		// Create the deps.js file for our project
		f, err := os.Create(path.Join(conf.Build, "deps.js"))
		if err != nil {
			return app.Error(err)
		}
		defer f.Close()

		// Write the deps file
		if err := WriteDeps(f, deps); err != nil {
			return err
		}

		// Compile the javascript
		if err := JsCompiler(out, deps); err != nil {
			return err
		}
	}

	log.Println("Done compiling! Elapsed:", time.Since(start))

	return nil
}

func JsCompiler(out string, deps []*Source) error {
	conf := config.Current()

	// Prepare the call to the compiler
	args := []string{
		"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--js_output_file", out,
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "base.js"),
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "deps.js"),
		"--js", path.Join(conf.Build, "deps.js"),
		"--output_wrapper", "(function(){%output%})();",
	}

	if conf.RootGss != "" {
		args = append(args, "--js", path.Join(conf.Build, "renaming-map.js"))
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
		args = append(args, "--jscomp_"+strings.ToLower(check), k)
	}

	// Add the compilation mode
	if conf.Mode == "ADVANCED" {
		args = append(args, "--compilation_level", "ADVANCED_OPTIMIZATIONS")
	} else if conf.Mode == "SIMPLE" {
		args = append(args, "--compilation_level", "SIMPLE_OPTIMIZATIONS")
	} else if conf.Mode == "WHITESPACE" {
		args = append(args, "--compilation_level", "WHITESPACE_ONLY")
	}

	// Add the warning level
	args = append(args, "--warning_level", conf.Level)

	// Add the externs
	for _, extern := range conf.Externs {
		args = append(args, "--externs", extern)
	}

	// Output the command that we'll run.
	if *outputCmd {
		f, err := os.Create(path.Join(conf.Build, "cmd"))
		if err != nil {
			return app.Error(err)
		}
		fmt.Fprintln(f, args)
		f.Close()
	}

	log.Println("Compiling code to build/compiled.js")

	// Compile the code
	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return app.Errorf("exec error: %s\n%s", err, string(output))
	}

	// If the compiler outputs something, send it to the console
	// for logging (so don't clubber the JS output of the handler).
	if len(output) > 0 {
		log.Println("Output from compiler:\n", string(output))
	}

	return nil
}

func copyCssFile() error {
	conf := config.Current()

	src, err := os.Open(filepath.Join(conf.Build, gss.CSS_NAME))
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(*cssOutput)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}
