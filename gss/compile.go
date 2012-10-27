package gss

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/scan"
)

// Compiles the .gss files
func Compile() error {
	conf := config.Current()

	// Output early if there's no GSS files.
	if conf.RootGss == "" {
		// Create/Clean the renaming map file to avoid compilation errors (the JS
		// compiler assumes there's a file with this name there).
		f, err := os.Create(path.Join(conf.Build, config.RENAMING_MAP_NAME))
		if err != nil {
			return app.Error(err)
		}
		f.Close()

		return nil
	}

	gss, err := scan.Do(conf.RootGss, ".gss")
	if err != nil {
		return err
	}

	// No results, no compiling
	if len(gss) == 0 {
		// Create/Clean the renaming map file to avoid compilation errors (the JS
		// compiler assumes there's a file with this name there).
		f, err := os.Create(path.Join(conf.Build, config.RENAMING_MAP_NAME))
		if err != nil {
			return app.Error(err)
		}
		f.Close()

		return nil
	}

	// Check if the cached version is still ok
	modified := false
	for _, filepath := range gss {
		if m, err := cache.Modified("compile", filepath); err != nil {
			return err
		} else if m {
			modified = true
		}
	}

	if !modified && !config.Build {
		return nil
	}

	log.Println("Compiling GSS...")

	// Run the soy compiler
	cmd := exec.Command(
		"java",
		"-jar", path.Join(conf.ClosureStylesheets, "build", "closure-stylesheets.jar"),
		"--output-file", filepath.Join(conf.Build, config.CSS_NAME),
		"--output-renaming-map-format", "CLOSURE_COMPILED",
		"--rename", "CLOSURE",
		"--output-renaming-map", path.Join(conf.Build, config.RENAMING_MAP_NAME))
	cmd.Args = append(cmd.Args, gss...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) != 0 {
			fmt.Println(string(output))
			os.Exit(1)
		}

		return app.Errorf("exec error: %s", err)
	}

	log.Println("Done compiling GSS!")

	return nil
}
