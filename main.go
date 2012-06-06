package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"code.google.com/p/gorilla/mux"
)

var (
	port      = flag.String("port", ":9810", "the port where the server will be listening")
	confArg   = flag.String("conf", "", "the config file")
	outputCmd = flag.Bool("output-cmd", false, "output compiler command to a file")
	build     = flag.Bool("build", false, "build the compiled files only and exit")
	cssOutput = flag.String("css-output", "compiled.css", "the css file that will be built")
	jsOutput  = flag.String("js-output", "compiled.js", "the js file that will be built")
)

func main() {
	flag.Parse()

	if err := ReadConf(); err != nil {
		log.Fatal(err)
	}

	if *build {
		Build()
	} else {
		Serve()
	}
}

func Serve() {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)
	addHandlers(r)

	log.Printf("Started closurer server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func addHandlers(r *mux.Router) {
	r.Handle("/", Handler(HomeHandler))
	r.Handle("/compile", Handler(CompileHandler))
	r.Handle("/css", Handler(CompileCssHandler))
}

func Build() {
	if err := CompileJs(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
