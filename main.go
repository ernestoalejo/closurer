package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/test"

	"github.com/gorilla/mux"
)

var exitServer = make(chan bool)

func main() {
	flag.Parse()

	if config.BuildTargets == "" {
		fmt.Println("Target required")
		flag.Usage()
		return
	}

	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	if config.Build {
		for _, t := range config.TargetList() {
			config.SelectTarget(t)

			if err := build(); err != nil {
				err.(*app.AppError).Log()
				break
			}
		}
	} else {
		if len(config.TargetList()) != 1 {
			log.Fatal("Cannot serve more than one target at the same time")
		}
		config.SelectTarget(config.TargetList()[0])

		serve()
	}
}

func serve() {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	r.Handle("/", app.Handler(home))
	r.Handle("/compile", app.Handler(compile))
	r.Handle("/input/{name:.+}", app.Handler(Input))
	r.Handle("/test/all", app.Handler(test.TestAll))
	r.Handle("/test/list", app.Handler(test.TestList))
	r.Handle("/test/{name:.+}", app.Handler(test.Main))
	r.Handle("/exit", app.Handler(exit))

	log.Printf("Started closurer server on http://localhost%s/\n", config.Port)
	go http.ListenAndServe(config.Port, nil)
	<-exitServer
}

func home(r *app.Request) error {
	return r.ExecuteTemplate([]string{"home"}, nil)
}

func exit(r *app.Request) error {
	exitServer <- true
	return nil
}

func compile(r *app.Request) error {
	conf := config.Current()
	target := conf.Js.CurTarget()

	if target.Mode == "RAW" {
		return RawOutput(r)
	}

	return js.CompiledJs(r)
}
