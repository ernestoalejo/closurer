package config

import (
	"flag"
	"strings"
)

var (
	// Command line flags
	Build, NoCache, OutputCmd    bool
	Port, ConfPath, BuildTargets string
)

var (
	SelectedTarget string
)

func init() {
	flag.BoolVar(&Build, "build", false, "build the compiled files only and exit")
	flag.BoolVar(&NoCache, "no-cache", false, "disables the files cache")
	flag.BoolVar(&OutputCmd, "output-cmd", false, "output compiler issued command to a file")
	flag.StringVar(&ConfPath, "conf", "", "the config file")
	flag.StringVar(&Port, "port", ":9810", "the port where the server will be listening")
	flag.StringVar(&BuildTargets, "targets", "", "the targets to run/compile, separated by colon")
}

func IsTarget(name string) bool {
	for _, t := range TargetList() {
		if t == name {
			return true
		}
	}
	return false
}

func TargetList() []string {
	return strings.Split(BuildTargets, ",")
}

func SelectTarget(target string) {
	SelectedTarget = target
}
