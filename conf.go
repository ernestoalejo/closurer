package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Id string `json:"id"`

	Root  string   `json:"root"`
	Paths []string `json:"paths"`
	Build string   `json:"build"`

	ClosureLibrary   string `json:"closure-library"`
	ClosureCompiler  string `json:"closure-compiler"`
	ClosureTemplates string `json:"closure-templates"`

	Mode string `json:"mode"`

	Inputs []string `json:"inputs"`

	Checks map[string]string `json:"checks"`
	Define map[string]string `json:"define"`
}

var conf = new(Config)

func ReadConf(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(conf); err != nil {
		return err
	}

	log.Println("Read app config: ", conf.Id)

	return nil
}
