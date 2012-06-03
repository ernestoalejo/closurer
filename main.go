package main

import (
	"flag"
	"log"
	"net/http"

	"code.google.com/p/gorilla/mux"
)

var (
	port      = flag.String("port", ":9810", "the port where the server will be listening")
	confArg   = flag.String("conf", "", "the config file")
	outputCmd = flag.Bool("output-cmd", false, "output compiler command to a file")
)

func main() {
	flag.Parse()

	if err := ReadConf(); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)
	addHandlers(r)

	log.Printf("Started closurer server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func addHandlers(r *mux.Router) {
	r.Handle("/", Handler(HomeHandler))
	r.Handle("/compile", Handler(CompileHandler))
}
