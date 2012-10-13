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
	for _, input := range conf.Inputs {
		if strings.Contains(input, "_test") {
			continue
		}

		ns, err := depstree.GetProvides(input)
		if err != nil {
			return nil, nil, err
		}
		namespaces = append(namespaces, ns...)
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
