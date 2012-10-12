package config

import (
	"flag"

	"github.com/ernestokarim/closurer/cache"
)

var (
	Build          bool
	Port, ConfPath string
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
	flag.StringVar(&Port, "port", ":9810", "the port where the server will be listening")
	flag.StringVar(&ConfPath, "conf", "", "the config file")

	flag.BoolVar(&cache.NoCache, "no-cache", false, "disables the files cache")
}
