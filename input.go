package main

import (
	"io"
	"os"
	"path"

	"github.com/gorilla/mux"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/hooks"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/soy"
)

func Input(r *app.Request) error {
	name := mux.Vars(r.Req)["name"]

	if err := hooks.PreCompile(); err != nil {
		return err
	}

	if name == config.DEPS_NAME {
		if err := soy.Compile(); err != nil {
			return err
		}

		if _, _, err := js.GenerateDeps("input"); err != nil {
			return err
		}

		conf := config.Current()
		f, err := os.Open(path.Join(conf.Build, config.DEPS_NAME))
		if err != nil {
			return app.Error(err)
		}
		defer f.Close()

		r.W.Header().Set("Content-Type", "text/javascript")
		if _, err := io.Copy(r.W, f); err != nil {
			return app.Error(err)
		}

		if err := hooks.PreCompile(); err != nil {
			return err
		}

		return nil
	}

	// Otherwise serve the file if it can be found
	paths := scan.BaseJSPaths()
	for _, p := range paths {
		f, err := os.Open(path.Join(p, name))
		if err != nil && !os.IsNotExist(err) {
			return app.Error(err)
		} else if err == nil {
			defer f.Close()

			r.W.Header().Set("Content-Type", "text/javascript")
			io.Copy(r.W, f)

			// Execute the post-compile actions
			if err := hooks.PostCompile(); err != nil {
				return err
			}

			return nil
		}
	}

	return app.Errorf("file not found: %s", name)
}
