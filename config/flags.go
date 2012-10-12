package config

import (
	"flag"
)

var (
	Build bool
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
}
