package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

func CompileHandler(r *Request) error {
	if err := CompileTemplates(r); err != nil {
		return err
	}

	if err := GenerateDeps(r); err != nil {
		return err
	}

	return nil
}

func GenerateDeps(r *Request) error {
	depswriter := path.Join(conf.ClosureLibrary, "closure", "bin", "build", "depswriter.py")
	output_file := path.Join(conf.Build, "deps.js")
	soyutils := path.Join(conf.ClosureTemplates, "javascript", "soyutils_usegoog.js")
	templates := path.Join(conf.Build, "templates")

	roots := []string{PrepareRoot(conf.Root), PrepareRoot(templates)}
	for _, root := range conf.Paths {
		roots = append(roots, PrepareRoot(root))
	}

	cmd := exec.Command("python", depswriter, "--output_file="+output_file)
	cmd.Args = append(cmd.Args, roots...)
	cmd.Args = append(cmd.Args, soyutils)

	log.Println("Writing dependencies in build/deps.js")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, "cannot generate the dependencies")
	}

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

	return nil
}
