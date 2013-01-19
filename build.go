package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/js"
)

var (
	mapping = map[string]string{}
)

func build() error {
	if err := js.FullCompile(); err != nil {
		return err
	}

	if err := copyCssFile(); err != nil {
		return err
	}

	if err := copyJsFile(); err != nil {
		return err
	}

	if err := outputMap(); err != nil {
		return err
	}

	return nil
}

func copyCssFile() error {
	conf := config.Current()
	target := conf.Gss.CurTarget()

	if conf.Gss == nil {
		return nil
	}

	srcName := filepath.Join(conf.Build, config.CSS_NAME)
	filename := target.Output
	if strings.Contains(filename, "{sha1}") {
		sha1, err := calcFileSha1(srcName)
		if err != nil {
			return err
		}
		filename = strings.Replace(filename, "{sha1}", sha1, -1)
	}

	mapping[config.SelectedTarget+"-css"] = filename

	if err := copyFile(srcName, filename); err != nil {
		return err
	}

	return nil
}

func copyJsFile() error {
	conf := config.Current()
	target := conf.Js.CurTarget()

	if conf.Js == nil {
		return nil
	}

	srcName := filepath.Join(conf.Build, config.JS_NAME)

	filename := filepath.Join(conf.Js.Root, target.Output)
	if strings.Contains(filename, "{sha1}") {
		sha1, err := calcFileSha1(srcName)
		if err != nil {
			return err
		}
		filename = strings.Replace(filename, "{sha1}", sha1, -1)
	}

	mapping[config.SelectedTarget+"-js"] = filename

	files := []string{}
	for _, n := range conf.Js.Prepends {
		files = append(files, filepath.Join(conf.Js.Root, n.File))
	}
	files = append(files, srcName)

	if err := copyFiles(files, filename); err != nil {
		return err
	}

	return nil
}

func calcFileSha1(filename string) (string, error) {
	cmd := exec.Command("sha1sum", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", app.Error(err)
	}

	return strings.Split(string(output), " ")[0], nil
}

func outputMap() error {
	conf := config.Current()
	if conf.Map == nil {
		return nil
	}

	f, err := os.Create(conf.Map.File)
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	fmt.Fprintf(f, "var mapping = ")
	if err := json.NewEncoder(f).Encode(&mapping); err != nil {
		return app.Error(err)
	}

	return nil
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return app.Error(err)
	}
	defer src.Close()

	dest, err := os.Create(to)
	if err != nil {
		return app.Error(err)
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return app.Error(err)
	}

	return nil
}

func copyFiles(from []string, to string) error {
	srcs := []io.Reader{}
	for _, f := range from {
		src, err := os.Open(f)
		if err != nil {
			return app.Error(err)
		}
		defer src.Close()

		srcs = append(srcs, src)
	}

	src := io.MultiReader(srcs...)

	dest, err := os.Create(to)
	if err != nil {
		return app.Error(err)
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return app.Error(err)
	}

	return nil
}
