package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
)

var (
	port       = flag.String("port", ":9810", "the port where the server will be listening")
	confArg    = flag.String("conf", "", "the config file")
	outputCmd  = flag.Bool("output-cmd", false, "output compiler command to a file")
	build      = flag.Bool("build", false, "build the compiled files only and exit")
	cssOutput  = flag.String("css-output", "compiled.css", "the css file that will be built")
	jsOutput   = flag.String("js-output", "compiled.js", "the js file that will be built")
	bench      = flag.Bool("bench", false, "enables internal circuits for benchmarks")
	cpuProfile = flag.String("cpu-profile", "", "write cpu profile to file")
	memProfile = flag.String("mem-profile", "", "write memory profile to file")
	noCache    = flag.Bool("no-cache", false, "disables the files cache")
)

func main() {
	flag.Parse()

	// Read configuration
	if err := ReadConf(); err != nil {
		log.Fatal(err)
	}

	// Read caches
	if !*noCache {
		if err := ReadDepsCache(); err != nil {
			log.Fatal(err)
		}
		if err := ReadSoyCache(); err != nil {
			log.Fatal(err)
		}
	}

	// Performs the benchmarks
	if *bench {
		if err := Bench(); err != nil {
			log.Fatal(err)
		}
		return
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
	http.Handle("/test/all", Handler(TestAllHandler))

	log.Printf("Started closurer server on http://localhost%s/\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func Build() {
	if err := CompileJs(os.Stdout); err != nil {
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

	for i := 0; i < 10; i += 1 {
		log.Println("Loop:", i)

		// Build the deps tree
		depstree, err := BuildDepsTree()
		if err != nil {
			return err
		}

		// Calculate all the input namespaces
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

		// Calculate the list of files to compile
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
