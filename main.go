package main

import (
	"flag"
	"log"
	"net/http"
	"os"
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

	// Read configuration
	if err := ReadConf(); err != nil {
		log.Fatal(err)
	}

	// Read caches
	if err := ReadDepsCache(); err != nil {
		log.Fatal(err)
	}
	if err := ReadSoyCache(); err != nil {
		log.Fatal(err)
	}

	// Start the correct mode
	if *build {
		Build()
	} else {
		Serve()
	}
}

func Serve() {
	http.Handle("/", Handler(HomeHandler))
	http.Handle("/compile", Handler(CompileHandler))
	http.Handle("/css", Handler(CompileCssHandler))
	http.Handle("/input/", Handler(InputHandler))
	http.Handle("/test/", Handler(TestHandler))

	log.Printf("Started closurer server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func Build() {
	if err := CompileJs(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
