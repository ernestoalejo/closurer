package gss

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
)

// Compiles the .gss files
func Compile() error {
	conf := config.Current()
	target := conf.Gss.CurTarget()

	// Output early if there's no GSS files.
	if conf.Gss == nil {
		if err := cleanRenamingMap(); err != nil {
			return err
		}

		return nil
	}

	// Check if the cached version is still ok
	modified := false
	for _, input := range conf.Gss.Inputs {
		if m, err := cache.Modified("compile", input.File); err != nil {
			return err
		} else if m {
			modified = true
			break
		}
	}

	if !modified {
		return nil
	}

	log.Println("Compiling GSS:", target.Name)

	if err := cleanRenamingMap(); err != nil {
		return err
	}

	// Prepare the list of non-standard functions.
	funcs := []string{}
	for _, f := range conf.Gss.Funcs {
		funcs = append(funcs, "--allowed-non-standard-function")
		funcs = append(funcs, f.Name)
	}

	// Prepare the renaming map args
	renaming := []string{}
	if target.Rename == "true" {
		renaming = []string{
			"--output-renaming-map-format", "CLOSURE_COMPILED",
			"--rename", "CLOSURE",
			"--output-renaming-map", path.Join(conf.Build, config.RENAMING_MAP_NAME),
		}
	}

	// Prepare the defines
	defines := []string{}
	for _, define := range target.Defines {
		defines = append(defines, "--define", define.Name)
	}

	// Prepare the inputs
	inputs := []string{}
	for _, input := range conf.Gss.Inputs {
		inputs = append(inputs, input.File)
	}

	// Prepare the command
	cmd := exec.Command(
		"java",
		"-jar", path.Join(conf.Gss.Compiler, "build", "closure-stylesheets.jar"),
		"--output-file", filepath.Join(conf.Build, config.CSS_NAME))
	cmd.Args = append(cmd.Args, funcs...)
	cmd.Args = append(cmd.Args, renaming...)
	cmd.Args = append(cmd.Args, inputs...)
	cmd.Args = append(cmd.Args, defines...)

	// Output the command if asked to
	if config.OutputCmd {
		fmt.Println("java", strings.Join(cmd.Args, " "))
	}

	// Run the compiler
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) != 0 {
			fmt.Println(string(output))
		}

		return app.Errorf("exec error: %s", err)
	}

	if len(output) > 0 {
		log.Println("Output from GSS compiler:\n", string(output))
	}

	log.Println("Done compiling GSS!")

	return nil
}

func cleanRenamingMap() error {
	conf := config.Current()

	// Create/Clean the renaming map file to avoid compilation errors (the JS
	// compiler assumes there's a file with this name there).
	f, err := os.Create(path.Join(conf.Build, config.RENAMING_MAP_NAME))
	if err != nil {
		return app.Error(err)
	}
	f.Close()

	return nil
}
