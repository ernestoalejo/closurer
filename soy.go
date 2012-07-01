package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Compile all modified templates
func CompileSoy() error {
	// Search the templates
	soy, err := Scan(conf.RootSoy, ".soy")
	if err != nil {
		return err
	}

	// No results, no compiling
	if len(soy) == 0 {
		return nil
	}

	for _, t := range soy {
		// Checks if the cached version is ok
		if modified, err := CacheModified(t); err != nil {
			return err
		} else if !modified {
			continue
		}

		// Relativize the path
		prel, err := filepath.Rel(conf.RootSoy, t)
		if err != nil {
			return fmt.Errorf("cannot put relative the path %s: %s", t, err)
		}

		// Creates all the necessary directories
		out := path.Join(conf.Build, "templates", prel+".js")
		if err := os.MkdirAll(path.Dir(out), 0755); err != nil {
			return fmt.Errorf("cannot create the build tree: %s", out)
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
			return fmt.Errorf("soy compiler error for file %s: %s\n%s", t,
				err, string(output))
		}
	}

	return nil
}
