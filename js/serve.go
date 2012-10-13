package js

import (
	"io"
	"os"
	"path"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
)

func CompiledJs(r *app.Request) error {
	r.W.Header().Set("Content-Type", "text/javascript")
	conf := config.Current()

	if err := FullCompile(); err != nil {
		return err
	}

	f, err := os.Open(path.Join(conf.Build, config.JS_NAME))
	if err != nil {
		return app.Error(err)
	}
	defer f.Close()

	if _, err := io.Copy(r.W, f); err != nil {
		return app.Error(err)
	}

	return nil
}
