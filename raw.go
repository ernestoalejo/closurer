package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/soy"
)

func RawOutput(r *app.Request) error {
	log.Println("Output RAW mode")

	if err := gss.Compile(); err != nil {
		return err
	}

	if err := soy.Compile(); err != nil {
		return err
	}

	depstree, err := scan.NewDepsTree("input")
	if err != nil {
		return err
	}

	conf := config.Current()
	content := bytes.NewBuffer(nil)

	base := path.Join(conf.ClosureLibrary, "closure", "goog", "base.js")
	if err := AddFile(content, base); err != nil {
		return err
	}

	if err := AddFile(content, path.Join(conf.Build, "renaming-map.js")); err != nil {
		return err
	}

	namespaces := []string{}
	for _, input := range conf.Inputs {
		if strings.Contains(input, "_test") {
			continue
		}

		ns, err := depstree.GetProvides(input)
		if err != nil {
			return err
		}
		namespaces = append(namespaces, ns...)
	}

	deps, err := depstree.GetDependencies(namespaces)
	if err != nil {
		return err
	}

	if err := scan.WriteDeps(content, deps); err != nil {
		return err
	}

	data := map[string]interface{}{
		"Content":    template.HTML(string(content.Bytes())),
		"Port":       config.Port,
		"LT":         template.HTML("<"),
		"Namespaces": template.HTML("'" + strings.Join(namespaces, "', '") + "'"),
	}
	r.W.Header().Set("Content-Type", "text/javascript")
	return r.ExecuteTemplate([]string{"raw"}, data)
}

func AddFile(w io.Writer, name string) error {
	f, err := os.Open(name)
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	io.Copy(w, f)

	return nil
}
