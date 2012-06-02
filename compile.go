package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

var timesCache = map[string]time.Time{}

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

	log.Println("Done compiling! Elapsed:", time.Since(start))

	return nil
}

func PrepareRoot(name string) string {
	return "--root=" + name
}

func CompileTemplates(r *Request) error {
	templates, err := ScanTemplates(conf.Root)
	if err != nil {
		return InternalErr(err, "cannot scan the root directory")
	}

	for _, template := range templates {
		if err := CompileTemplate(r, template); err != nil {
			return err
		}
	}

	return nil
}

func ScanTemplates(filepath string) ([]string, error) {
	templates := []string{}

	ls, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil, err
	}

	for _, entry := range ls {
		fullpath := path.Join(filepath, entry.Name())

		if entry.IsDir() {
			t, err := ScanTemplates(fullpath)
			if err != nil {
				return nil, err
			}
			templates = append(templates, t...)
		} else if path.Ext(entry.Name()) == ".soy" {
			templates = append(templates, fullpath)
		}
	}

	return templates, nil
}

func CompileTemplate(r *Request, filepath string) error {
	soytojs := path.Join(conf.ClosureTemplates, "build", "SoyToJsSrcCompiler.jar")
	out := path.Join(conf.Build, "templates", filepath+".js")

	otime, ok := timesCache[filepath]
	if ok {
		info, err := os.Lstat(out)
		if err != nil && !os.IsNotExist(err) {
			return InternalErr(err, fmt.Sprintf("cannot check the file info: %s", out))
		}

		if info.ModTime() == otime {
			return nil
		}
	}

	if err := os.MkdirAll(path.Base(out), 0755); err != nil {
		return InternalErr(err, fmt.Sprintf("cannot create the build tree: %s", out))
	}

	cmd := exec.Command("java", "-jar", soytojs, "--outputPathFormat", out,
		"--shouldGenerateJsdoc", "--shouldProvideRequireSoyNamespaces",
		"--cssHandlingScheme", "goog", filepath)

	log.Println("Compiling template:", filepath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, fmt.Sprintf("cannot compile the template %s", filepath))
	}

	info, err := os.Lstat(out)
	if err != nil && !os.IsNotExist(err) {
		return InternalErr(err, fmt.Sprintf("cannot check the file info: %s", out))
	}
	timesCache[filepath] = info.ModTime()

	return nil
}

func CompileCode(r *Request, deps []*Source) error {
	args := []string{"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar")}

	for _, dep := range deps {
		args = append(args, "--js", dep.filename)
	}

	for k, define := range conf.Define {
		if define != "true" && define != "false" {
			define = "\"" + define + "\""
		}
		args = append(args, "--define", k+"="+define)
	}

	if conf.Mode == "ADVANCED" {
		args = append(args, "--compilation_level", "ADVANCED_OPTIMIZATIONS")
	} else if conf.Mode == "SIMPLE" {
		args = append(args, "--compilation_level", "SIMPLE_OPTIMIZATIONS")
	} else if conf.Mode == "WHITESPACE" {
		args = append(args, "--compilation_level", "WHITESPACE_ONLY")
	} else {
		return fmt.Errorf("compilation mode not recognized: %s", conf.Mode)
	}

	if conf.Level == "QUIET" || conf.Level == "DEFAULT" || conf.Level == "VERBOSE" {
		args = append(args, "--warning_level", conf.Level)
	} else {
		return fmt.Errorf("warnings level not recognized: %s", conf.Level)
	}

	log.Println("Compiling code to build/compiled.js")

	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, "cannot compile the code")
	}

	fmt.Fprintf(r.W, "%s\n", output)

	return nil
}
