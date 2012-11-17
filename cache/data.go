package cache

import (
	"github.com/ernestokarim/closurer/config"
)

var dataCache = map[string]interface{}{}

// Read some data of the cache with the key. If the data it's not present,
// blank will be returned.
func ReadData(key string, blank interface{}) interface{} {
	d, ok := dataCache[key]
	if !ok || config.NoCache {
		dataCache[key] = blank
		return blank
	}

	return d
}
