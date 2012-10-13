package config

import (
	"flag"
)

var (
	Build, NoCache bool
	Port, ConfPath string
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
	flag.BoolVar(&NoCache, "no-cache", false, "disables the files cache")
	flag.StringVar(&ConfPath, "conf", "", "the config file")
	flag.StringVar(&Port, "port", ":9810", "the port where the server will be listening")
}
