package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

var soyCache = map[string]time.Time{}

// Compile all modified templates
func CompileSoy(w io.Writer) error {
	templates, err := ScanTemplates(conf.RootSoy)
	if err != nil {
		return fmt.Errorf("cannot scan templates: %s", err)
	}

	for _, template := range templates {
		if err := SoyCompiler(w, template); err != nil {
			return err
		}
	}

	if err := WriteSoyCache(); err != nil {
		return err
	}

	return nil
}

// Scan a directory searching for .soy files
func ScanTemplates(filepath string) ([]string, error) {
	templates := []string{}

	// Get the list of entries
	ls, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil, err
	}

	for _, entry := range ls {
		fullpath := path.Join(filepath, entry.Name())

		if entry.IsDir() {
			if IsValidDir(entry.Name()) {
				// Scan recursively the directories
				t, err := ScanTemplates(fullpath)
				if err != nil {
					return nil, err
				}
				templates = append(templates, t...)
			}
		} else if path.Ext(entry.Name()) == ".soy" {
			// Add the templates to the list
			templates = append(templates, fullpath)
		}
	}

	return templates, nil
}

// Compile a template if it has been modified
func SoyCompiler(w io.Writer, p string) error {
	prel, err := filepath.Rel(conf.RootSoy, p)
	if err != nil {
		return fmt.Errorf("cannot relativize the path to %s: %s", p, err)
	}

	soytojs := path.Join(conf.ClosureTemplates, "build", "SoyToJsSrcCompiler.jar")
	out := path.Join(conf.Build, "templates", prel+".js")

	// Get the stat file info
	info, err := os.Lstat(p)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot check the file info: %s", p)
	}

	// Check if the cached version is still ok
	otime, ok := soyCache[p]
	if ok {
		if info.ModTime() == otime {
			return nil
		}
	}

	// Creates all the necessary directories
	if err := os.MkdirAll(path.Dir(out), 0755); err != nil {
		return fmt.Errorf("cannot create the build tree: %s", out)
	}

	log.Println("Compiling template:", p)

	// Compile the template
	cmd := exec.Command("java", "-jar", soytojs, "--outputPathFormat", out,
		"--shouldGenerateJsdoc", "--shouldProvideRequireSoyNamespaces",
		"--cssHandlingScheme", "goog", p)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "%s\n", output)
		return fmt.Errorf("cannot compile the template %s: %s", p, err)
	}

	// Cache the output
	soyCache[p] = info.ModTime()

	return nil
}

func ReadSoyCache() error {
	name := path.Join(conf.Build, "soy-cache")
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	log.Println("Reading soy cache:", name)

	if err := gob.NewDecoder(f).Decode(&soyCache); err != nil {
		return fmt.Errorf("cannot decode the deps cache: %s", err)
	}

	return nil
}

func WriteSoyCache() error {
	f, err := os.Create(path.Join(conf.Build, "soy-cache"))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(&soyCache); err != nil {
		return fmt.Errorf("cannot encode the deps cache: %s", err)
	}

	return nil
}
