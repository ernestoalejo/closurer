package config

import (
	"flag"
)

var (
	Build, NoCache, OutputCmd bool
	Port, ConfPath            string
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
	flag.BoolVar(&NoCache, "no-cache", false, "disables the files cache")
	flag.BoolVar(&OutputCmd, "output-cmd", false, "output compiler issued command to a file")
	flag.StringVar(&ConfPath, "conf", "", "the config file")
	flag.StringVar(&Port, "port", ":9810", "the port where the server will be listening")
}
