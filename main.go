package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/ernestokarim/closurer/cache"
	"github.com/ernestokarim/closurer/config"
)

var (
	port       = flag.String("port", ":9810", "the port where the server will be listening")
	confArg    = flag.String("conf", "", "the config file")
	outputCmd  = flag.Bool("output-cmd", false, "output compiler command to a file")
	cssOutput  = flag.String("css-output", "compiled.css", "the css file that will be built")
	jsOutput   = flag.String("js-output", "compiled.js", "the js file that will be built")
	bench      = flag.Bool("bench", false, "enables internal circuits for benchmarks")
	cpuProfile = flag.String("cpu-profile", "", "write cpu profile to file")
	memProfile = flag.String("mem-profile", "", "write memory profile to file")
	noCache    = flag.Bool("no-cache", false, "disables the files cache")
)

func main() {
	flag.Parse()

	cache.SetNoCache(*noCache)

	if err := config.ReadFromFile(*confArg); err != nil {
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
	http.Handle("/", Handler(HomeHandler))
	http.Handle("/compile", Handler(CompileHandler))
	//http.Handle("/css", Handler(CompileGssHandler))
	http.Handle("/input/", Handler(InputHandler))
	http.Handle("/test/", Handler(TestHandler))
	http.Handle("/test/all", Handler(TestAllHandler))
	http.Handle("/test/list", Handler(TestListHandler))

	log.Printf("Started closurer server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func Build() {
	if err := PreCompileActions(); err != nil {
		log.Fatal(err)
	}

	if err := CompileJs(os.Stdout); err != nil {
		log.Fatal(err)
	}

	if err := PostCompileActions(); err != nil {
		log.Fatal(err)
	}
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

		depstree, err := NewDepsTree("bench")
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
