package cache

import (

)

var dataCache = map[string]interface{}{}

// Write a new symbol with its key and value in the cache.
func WriteData(key string, value interface{}) {
	dataCache[key] = value
}

// Read some data of the cache with the key. If the data it's not present,
// blank will be returned.
func ReadData(key string, blank interface{}) interface{} {
	d, ok := dataCache[key]
	if !ok {
		dataCache[key] = blank
		return blank
	}

	return d
}

