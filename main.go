package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/test"

	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()

	if err := config.ReadFromFile(config.ConfPath); err != nil {
		log.Fatal(err)
	}

	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	if config.Build {
		if err := build(); err != nil {
			log.Fatal(err)
		}
	} else {
		serve()
	}
}

func serve() {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	r.Handle("/", app.Handler(home))
	r.Handle("/compile", app.Handler(compile))
	r.Handle("/input/{name:.+}", app.Handler(Input))
	r.Handle("/test", app.Handler(test.Main))
	r.Handle("/test/all", app.Handler(test.TestAll))
	r.Handle("/test/list", app.Handler(test.TestList))

	log.Printf("Started closurer server on http://localhost%s/\n", config.Port)
	log.Fatal(http.ListenAndServe(config.Port, nil))
}

func home(r *app.Request) error {
	return r.ExecuteTemplate([]string{"home"}, nil)
}

func compile(r *app.Request) error {
	conf := config.Current()
	if conf.Mode == "RAW" {
		return RawOutput(r)
	}

	return js.CompiledJs(r)
}
