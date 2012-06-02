package main

import (
	"flag"
	"log"
	"net/http"

	"code.google.com/p/gorilla/mux"
)

var (
	port    = flag.String("port", ":9810", "the port where the server will be listening")
	confArg = flag.String("conf", "", "the config file")
)

func main() {
	flag.Parse()

	if err := ReadConf(*confArg); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)
	addHandlers(r)

	log.Printf("Started closure server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func addHandlers(r *mux.Router) {
	r.Handle("/", Handler(HomeHandler))
}
