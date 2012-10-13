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
	"github.com/ernestokarim/closurer/utils"
)

// Compile all modified templates
func Compile() error {
	conf := config.Current()

	if conf.RootSoy == "" {
		return nil
	}

	soy, err := utils.Scan(conf.RootSoy, ".soy")
	if err != nil {
		return err
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

		prel, err := filepath.Rel(conf.RootSoy, t)
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

		log.Println("Done compiling template!")
	}

	return nil
}
