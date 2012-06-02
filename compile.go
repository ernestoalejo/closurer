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

var timesCache = map[string]time.Time{}

func CompileHandler(r *Request) error {
	if err := CompileTemplates(r); err != nil {
		return err
	}

	if err := CompileCode(r); err != nil {
		return err
	}

	f, err := os.Open(path.Join(conf.Build, "compiled.js"))
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(r.W, f)

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

func CompileCode(r *Request) error {
	closurebuilder := path.Join(conf.ClosureLibrary, "closure", "bin", "build",
		"closurebuilder.py")
	output_file := path.Join(conf.Build, "compiled.js")
	soyutils := path.Join(conf.ClosureTemplates, "javascript")
	templates := path.Join(conf.Build, "templates")
	closure_library := path.Join(conf.ClosureLibrary)

	roots := []string{
		PrepareRoot(templates), PrepareRoot(conf.Root),
		PrepareRoot(closure_library), PrepareRoot(soyutils),
	}
	for _, root := range conf.Paths {
		roots = append(roots, PrepareRoot(root))
	}

	inputs := []string{}
	for _, input := range conf.Inputs {
		inputs = append(inputs, "--input")
		inputs = append(inputs, input)
	}

	cmd := exec.Command("python", closurebuilder, "--output_file="+output_file,
		"--compiler_jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--output_mode", "compiled")
	cmd.Args = append(cmd.Args, roots...)
	cmd.Args = append(cmd.Args, inputs...)

	if conf.Mode == "ADVANCED" {
		cmd.Args = append(cmd.Args, "--compiler_flags")
		cmd.Args = append(cmd.Args, "--compilation_level=ADVANCED_OPTIMIZATIONS")
	} else if conf.Mode == "SIMPLE" {
		cmd.Args = append(cmd.Args, "--compiler_flags")
		cmd.Args = append(cmd.Args, "--compilation_level=SIMPLE_OPTIMIZATIONS")
	} else {
		return fmt.Errorf("compilation mode not recognized: %s", conf.Mode)
	}

	log.Println("Compiling code to build/compiled.js")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, "cannot compile the code")
	}

	return nil
}
