package js

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/domain"
	"github.com/ernestokarim/closurer/scan"
)

func GenerateDeps(dest string) ([]*domain.Source, []string, error) {
	log.Println("Scanning deps...")

	conf := config.Current()

	depstree, err := scan.NewDepsTree(dest)
	if err != nil {
		return nil, nil, err
	}

	namespaces := []string{}
	for _, input := range conf.Js.Inputs {
		if dest != "input" && strings.Contains(input.File, "_test") {
			continue
		}

		ns, err := depstree.GetProvides(input.File)
		if err != nil {
			return nil, nil, err
		}
		namespaces = append(namespaces, ns...)
	}

	if dest == "input" {
		// Add the necesary namespaces for the multi-test runner
		namespaces = append(namespaces, "goog.style")
		namespaces = append(namespaces, "goog.userAgent.product")
		namespaces = append(namespaces, "goog.testing.MultiTestRunner")
		namespaces = append(namespaces, depstree.GetTestingNamespaces()...)
	} else if dest == "input-production" {
		namespaces = append(namespaces, "goog.style")
	}

	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return nil, nil, err
	}

	f, err := os.Create(filepath.Join(conf.Build, config.DEPS_NAME))
	if err != nil {
		return nil, nil, app.Error(err)
	}
	defer f.Close()

	if err := scan.WriteDeps(f, deps); err != nil {
		return nil, nil, err
	}

	log.Println("Done scanning deps!")

	return deps, namespaces, nil
}
