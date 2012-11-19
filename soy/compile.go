package soy

import (
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

// Compile all modified templates
func Compile() error {
	conf := config.Current()

	if conf.Soy == nil || conf.Soy.Root == "" {
		return nil
	}

	if err := os.MkdirAll(path.Join(conf.Build, "templates"), 0755); err != nil {
		return app.Error(err)
	}

	buildPrefix := filepath.Join(conf.Build, "templates")
	oldSoy, err := scan.Do(buildPrefix, ".js")
	if err != nil {
		return err
	}

	soy, err := scan.Do(conf.Soy.Root, ".soy")
	if err != nil {
		return err
	}

	indexed := map[string]bool{}
	for _, f := range soy {
		f = f[len(conf.Soy.Root):]
		indexed[f] = true
	}

	// Delete compiled templates no longer present in the sources
	for _, f := range oldSoy {
		compare := f[len(buildPrefix) : len(f)-3]
		if _, ok := indexed[compare]; !ok {
			if err := os.Remove(f); err != nil {
				return app.Error(err)
			}
		}
	}

	if len(soy) == 0 {
		return nil
	}

	for _, t := range soy {
		if modified, err := cache.Modified("compile", t); err != nil {
			return err
		} else if !modified {
			continue
		}

		prel, err := filepath.Rel(conf.Soy.Root, t)
		if err != nil {
			return app.Error(err)
		}

		out := path.Join(conf.Build, "templates", prel+".js")
		if err := os.MkdirAll(path.Dir(out), 0755); err != nil {
			return app.Error(err)
		}

		log.Println("Compiling template", t, "...")

		// Run the compiler command
		cmd := exec.Command(
			"java",
			"-jar", path.Join(conf.Soy.Compiler, "build", "SoyToJsSrcCompiler.jar"),
			"--outputPathFormat", out,
			"--shouldGenerateJsdoc",
			"--shouldProvideRequireSoyNamespaces",
			"--cssHandlingScheme", "goog",
			t)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return app.Errorf("exec error with %s: %s\n%s", t, err, string(output))
		}

		log.Println("Done compiling template!")
	}

	return nil
}
