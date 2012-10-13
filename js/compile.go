package js

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/domain"
)

func Compile(out string, deps []*domain.Source) error {
	conf := config.Current()

	args := []string{
		"-jar", path.Join(conf.ClosureCompiler, "build", "compiler.jar"),
		"--js_output_file", out,
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "base.js"),
		"--js", path.Join(conf.ClosureLibrary, "closure", "goog", "deps.js"),
		"--js", path.Join(conf.Build, "deps.js"),
		"--output_wrapper", "(function(){%output%})();",
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
