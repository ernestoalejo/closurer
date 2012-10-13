package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/utils"
)

// Compile all modified templates
func CompileSoy() error {
	conf := config.Current()

	// Output early if there's no SOY files
	if conf.RootSoy == "" {
		return nil
	}

	// Search the templates
	soy, err := utils.Scan(conf.RootSoy, ".soy")
	if err != nil {
		return err
	}

	// No results, no compiling
	if len(soy) == 0 {
		return nil
	}

	for _, t := range soy {
		// Checks if the cached version is ok
		if modified, err := cache.Modified("compile", t); err != nil {
			return err
		} else if !modified {
			continue
		}

		// Relativize the path
		prel, err := filepath.Rel(conf.RootSoy, t)
		if err != nil {
			return app.Error(err)
		}

		// Creates all the necessary directories
		out := path.Join(conf.Build, "templates", prel+".js")
		if err := os.MkdirAll(path.Dir(out), 0755); err != nil {
			return app.Error(err)
		}

		log.Println("Compiling template:", t)

		// Run the compiler command
		cmd := exec.Command(
			"java",
			"-jar", path.Join(conf.ClosureTemplates, "build", "SoyToJsSrcCompiler.jar"),
			"--outputPathFormat", out,
			"--shouldGenerateJsdoc",
			"--shouldProvideRequireSoyNamespaces",
			"--cssHandlingScheme", "goog",
			t)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return app.Errorf("exec error with %s: %s\n%s", t, err, string(output))
		}
	}

	return nil
}
