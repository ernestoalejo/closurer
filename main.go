package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/test"

	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()

	if err := config.ReadFromFile(config.ConfPath); err != nil {
		log.Fatal(err)
	}

	if config.Build {
		if err := Build(); err != nil {
			log.Fatal(err)
		}
	} else {
		Serve()
	}
}

func Serve() {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	r.Handle("/", app.Handler(Home))
	r.Handle("/compile", app.Handler(Compile))
	r.Handle("/css", app.Handler(gss.CompiledCss))
	r.Handle("/input/{name:.+}", app.Handler(Input))
	r.Handle("/test", app.Handler(test.Main))
	r.Handle("/test/all", app.Handler(test.TestAll))
	r.Handle("/test/list", app.Handler(test.TestList))

	log.Printf("Started closurer server on http://localhost%s/\n", config.Port)
	log.Fatal(http.ListenAndServe(config.Port, nil))
}

func Home(r *app.Request) error {
	return r.ExecuteTemplate([]string{"home"}, nil)
}

func Compile(r *app.Request) error {
	conf := config.Current()
	if conf.Mode == "RAW" {
		return RawOutput(r)
	}

	return js.CompiledJs(r)
}
