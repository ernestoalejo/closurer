package config

import (
	"flag"

	"github.com/ernestokarim/closurer/cache"
)

var (
	Build bool
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
	flag.BoolVar(&cache.NoCache, "no-cache", false, "disables the files cache")
}
