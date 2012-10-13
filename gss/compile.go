package gss

import (
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/utils"
)

const CSS_NAME = "compiled.css"

/*
// Returns the compiled CSS.
func CompileGss(r *app.Request) error {
	conf := config.Current()

	r.W.Header().Set("Content-Type", "text/css")

	// Output early if there's no GSS files
	if conf.RootGss == "" {
		fmt.Fprintln(r.W, "")
		return nil
	}

	// Execute the pre-compile actions
	if err := PreCompileActions(); err != nil {
		return err
	}

	// Compile the .gss files
	if err := CompileGss(); err != nil {
		return err
	}

	// Execute the post-compile actions
	if err := PostCompileActions(); err != nil {
		return err
	}

	// Copy the compile file to the output
	f, err := os.Open(path.Join(conf.Build, "compiled.css"))
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	io.Copy(r.W, f)

	return nil
}*/

// Compiles the .gss files
func Compile() error {
	conf := config.Current()

	// Create/Clean the renaming map file to avoid compilation errors (the JS
	// compiler assumes there's a file with this name there).
	f, err := os.Create(path.Join(conf.Build, "renaming-map.js"))
	if err != nil {
		return err
	}
	f.Close()

	// Output early if there's no GSS files.
	if conf.RootGss == "" {
		return nil
	}

	gss, err := utils.Scan(conf.RootGss, ".gss")
	if err != nil {
		return err
	}

	// No results, no compiling
	if len(gss) == 0 {
		return nil
	}

	// Check if the cached version is still ok
	modified := false
	for _, filepath := range gss {
		if m, err := cache.Modified("compile", filepath); err != nil {
			return err
		} else if m {
			modified = true
			break
		}
	}

	if !modified && !config.Build {
		return nil
	}

	log.Println("Compiling gss")

	// Compute the output path
	out := path.Join(conf.Build, CSS_NAME)

	// Run the soy compiler
	cmd := exec.Command(
		"java",
		"-jar", path.Join(conf.ClosureStylesheets, "build", "closure-stylesheets.jar"),
		"--output-file", out,
		"--output-renaming-map-format", "CLOSURE_COMPILED",
		"--rename", "CLOSURE",
		"--output-renaming-map", path.Join(conf.Build, "renaming-map.js"))
	cmd.Args = append(cmd.Args, gss...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return app.Errorf("exec error: %s\n%s", err, string(output))
	}

	return nil
}
