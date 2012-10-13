package js

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/domain"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/hooks"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/soy"
)

const JS_NAME = "compiled.js"

func FullCompile() error {
	if err := hooks.PreCompile(); err != nil {
		return err
	}

	if err := gss.Compile(); err != nil {
		return err
	}

	if err := soy.Compile(); err != nil {
		return err
	}

	if err := Compile(); err != nil {
		return err
	}

	if err := hooks.PostCompile(); err != nil {
		return err
	}

	return nil
}

func Compile() error {
	conf := config.Current()

	deps, err := generateDeps()
	if err != nil {
		return err
	}

	args := []string{
		"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--js_output_file", path.Join(conf.Build, JS_NAME),
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "base.js"),
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "deps.js"),
		"--js", filepath.Join(conf.Build, "deps.js"),
		"--output_wrapper", `(function(){%output%})();`,
	}

	for _, dep := range deps {
		if !strings.Contains(dep.Filename, "_test.js") {
			args = append(args, "--js", dep.Filename)
		}
	}

	for k, define := range conf.Define {
		if define != "true" && define != "false" {
			define = "\"" + define + "\""
		}
		args = append(args, "--define", k+"="+define)
	}

	for k, check := range conf.Checks {
		args = append(args, "--jscomp_"+strings.ToLower(check), k)
	}

	if conf.Mode == "ADVANCED" {
		args = append(args, "--compilation_level", "ADVANCED_OPTIMIZATIONS")
	} else if conf.Mode == "SIMPLE" {
		args = append(args, "--compilation_level", "SIMPLE_OPTIMIZATIONS")
	} else if conf.Mode == "WHITESPACE" {
		args = append(args, "--compilation_level", "WHITESPACE_ONLY")
	}

	args = append(args, "--warning_level", conf.Level)

	for _, extern := range conf.Externs {
		args = append(args, "--externs", extern)
	}

	if config.OutputCmd {
		f, err := os.Create(path.Join(conf.Build, "cmd"))
		if err != nil {
			return app.Error(err)
		}
		fmt.Fprintln(f, args)
		f.Close()
	}

	log.Println("Compiling JS...")

	cmd := exec.Command("java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return app.Errorf("exec error: %s\n%s", err, string(output))
	}

	if len(output) > 0 {
		log.Println("Output from compiler:\n", string(output))
	}

	log.Println("Done compiling JS!")

	return nil
}

func generateDeps() ([]*domain.Source, error) {
	conf := config.Current()

	depstree, err := scan.NewDepsTree("compile")
	if err != nil {
		return nil, err
	}

	namespaces := []string{}
	for _, input := range conf.Inputs {
		if strings.Contains(input, "_test") {
			continue
		}

		ns, err := depstree.GetProvides(input)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns...)
	}

	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(filepath.Join(conf.Build, "deps.js"))
	if err != nil {
		return nil, app.Error(err)
	}
	defer f.Close()

	if err := scan.WriteDeps(f, deps); err != nil {
		return nil, err
	}

	return deps, nil
}
