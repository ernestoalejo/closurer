package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Id string `json:"id"`
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
