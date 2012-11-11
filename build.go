package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/js"
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

	return nil
}

func copyCssFile() error {
	conf := config.Current()

	if conf.Gss.Root == "" {
		return nil
	}

	src, err := os.Open(filepath.Join(conf.Build, config.CSS_NAME))
	if err != nil {
		return app.Error(err)
	}
	defer src.Close()

	dest, err := os.Create(config.CssOutput)
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

	src, err := os.Open(filepath.Join(conf.Build, config.JS_NAME))
	if err != nil {
		return app.Error(err)
	}
	defer src.Close()

	dest, err := os.Create(config.JsOutput)
	if err != nil {
		return app.Error(err)
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return app.Error(err)
	}

	return nil
}
