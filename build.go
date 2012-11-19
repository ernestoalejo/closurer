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
	src, err := os.Open(srcName)
	if err != nil {
		return app.Error(err)
	}
	defer src.Close()

	filename := target.Output
	if strings.Contains(filename, "{sha1}") {
		sha1, err := calcFileSha1(srcName)
		if err != nil {
			return err
		}
		filename = strings.Replace(filename, "{sha1}", sha1, -1)
	}

	mapping[config.SelectedTarget+"-css"] = filename

	dest, err := os.Create(filename)
	if err != nil {
		return app.Error(err)
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return app.Error(err)
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
	src, err := os.Open(srcName)
	if err != nil {
		return app.Error(err)
	}
	defer src.Close()

	filename := target.Output
	if strings.Contains(filename, "{sha1}") {
		sha1, err := calcFileSha1(srcName)
		if err != nil {
			return err
		}
		filename = strings.Replace(filename, "{sha1}", sha1, -1)
	}

	mapping[config.SelectedTarget+"-js"] = filename

	dest, err := os.Create(filename)
	if err != nil {
		return app.Error(err)
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return app.Error(err)
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
