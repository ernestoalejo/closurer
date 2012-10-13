package hooks

import (
	"github.com/ernestokarim/closurer/cache"
)

// Called after each compilation tasks.
// It saves the caches.
func PostCompile() error {
	return cache.Dump()
}
