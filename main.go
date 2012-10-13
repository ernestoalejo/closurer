package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/gss"
	"github.com/ernestokarim/closurer/hooks"
	"github.com/ernestokarim/closurer/js"
	"github.com/ernestokarim/closurer/scan"
	"github.com/ernestokarim/closurer/test"

	"github.com/gorilla/mux"
)

var (
	cssOutput  = flag.String("css-output", "compiled.css", "the css file that will be built")
	jsOutput   = flag.String("js-output", "compiled.js", "the js file that will be built")
	bench      = flag.Bool("bench", false, "enables internal circuits for benchmarks")
	cpuProfile = flag.String("cpu-profile", "", "write cpu profile to file")
	memProfile = flag.String("mem-profile", "", "write memory profile to file")
)

func main() {
	flag.Parse()

	if err := config.ReadFromFile(config.ConfPath); err != nil {
		log.Fatal(err)
	}

	if *bench {
		if err := Bench(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if config.Build {
		Build()
	} else {
		Serve()
	}
}

func Serve() {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	r.Handle("/", app.Handler(Home))
	r.Handle("/compile", app.Handler(js.CompiledJs))
	r.Handle("/css", app.Handler(gss.CompiledCss))
	r.Handle("/input/{name:.+}", app.Handler(Input))
	r.Handle("/test", app.Handler(test.Main))
	r.Handle("/test/all", app.Handler(test.TestAll))
	r.Handle("/test/list", app.Handler(test.TestList))

	log.Printf("Started closurer server on http://localhost%s/\n", config.Port)
	log.Fatal(http.ListenAndServe(config.Port, nil))
}

func Build() {
	if err := hooks.PreCompile(); err != nil {
		log.Fatal(err)
	}

	if err := js.FullCompile(); err != nil {
		log.Fatal(err)
	}

	if err := hooks.PostCompile(); err != nil {
		log.Fatal(err)
	}

	if err := copyCssFile(); err != nil {
		log.Fatal(err)
	}
}

func copyCssFile() error {
	conf := config.Current()

	src, err := os.Open(filepath.Join(conf.Build, gss.CSS_NAME))
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(*cssOutput)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}

func Bench() error {
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			return err
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	conf := config.Current()

	for i := 0; i < 10; i += 1 {
		log.Println("Loop:", i)

		depstree, err := scan.NewDepsTree("bench")
		if err != nil {
			return err
		}

		namespaces := []string{}
		for _, input := range conf.Inputs {
			if strings.Contains(input, "_test") {
				continue
			}

			ns, err := depstree.GetProvides(input)
			if err != nil {
				return err
			}
			namespaces = append(namespaces, ns...)
		}

		_, err = depstree.GetDependencies(namespaces)
		if err != nil {
			return err
		}

		if *memProfile != "" {
			f, err := os.Create(*memProfile)
			if err != nil {
				return err
			}
			defer f.Close()

			pprof.WriteHeapProfile(f)

			return nil
		}
	}

	return nil
}
