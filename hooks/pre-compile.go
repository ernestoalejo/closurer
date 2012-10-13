package hooks

import (
	"os"
	"sync"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
)

var loadCacheOnce sync.Once

// Called before each compilation task. It load the caches
// and reload the confs if needed.
func PreCompile() error {
	if err := config.ReadFromFile(config.ConfPath); err != nil {
		return err
	}

	if err := config.Validate(); err != nil {
		return err
	}

	conf := config.Current()
	if err := os.MkdirAll(conf.Build, 0755); err != nil {
		return app.Error(err)
	}

	var err error
	loadCacheOnce.Do(func() {
		err = cache.Load()
	})

	return err
}
