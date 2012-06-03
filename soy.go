package main

import (
	"io/ioutil"
	"path"
	"os"
	"fmt"
	"os/exec"
	"time"
	"log"
)

var timesCache = map[string]time.Time{}

// Compile all modified templates
func CompileTemplates(r *Request) error {
	templates, err := ScanTemplates(conf.Root)
	if err != nil {
		return InternalErr(err, "cannot scan the root directory")
	}

	for _, template := range templates {
		if err := CompileTemplate(r, template); err != nil {
			return err
		}
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
			// Scan recursively the directories
			t, err := ScanTemplates(fullpath)
			if err != nil {
				return nil, err
			}
			templates = append(templates, t...)
		} else if path.Ext(entry.Name()) == ".soy" {
			// Add the templates to the list
			templates = append(templates, fullpath)
		}
	}

	return templates, nil
}

// Compile a template if it has been modified
func CompileTemplate(r *Request, filepath string) error {
	soytojs := path.Join(conf.ClosureTemplates, "build", "SoyToJsSrcCompiler.jar")
	out := path.Join(conf.Build, "templates", filepath+".js")

	// Get the stat file info
	info, err := os.Lstat(out)
	if err != nil && !os.IsNotExist(err) {
		return InternalErr(err, fmt.Sprintf("cannot check the file info: %s", out))
	}

	// Check if the cached version is still ok
	otime, ok := timesCache[filepath]
	if ok {
		if info.ModTime() == otime {
			return nil
		}
	}

	// Creates all the necessary directories
	if err := os.MkdirAll(path.Base(out), 0755); err != nil {
		return InternalErr(err, fmt.Sprintf("cannot create the build tree: %s", out))
	}

	log.Println("Compiling template:", filepath)

	// Compile the template
	cmd := exec.Command("java", "-jar", soytojs, "--outputPathFormat", out,
		"--shouldGenerateJsdoc", "--shouldProvideRequireSoyNamespaces",
		"--cssHandlingScheme", "goog", filepath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(r.W, "%s\n", output)
		return InternalErr(err, fmt.Sprintf("cannot compile the template %s", filepath))
	}

	// Cache the output
	info, err = os.Lstat(out)
	if err != nil && !os.IsNotExist(err) {
		return InternalErr(err, fmt.Sprintf("cannot check the file info: %s", out))
	}

	timesCache[filepath] = info.ModTime()

	return nil
}
