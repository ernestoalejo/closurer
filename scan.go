package main

import (
	"fmt"
	"io/ioutil"
	"path"
)

type visitor struct {
	results []string
}

func (v *visitor) scan(filepath string, ext string) error {
	// Get the list of entries
	ls, err := ioutil.ReadDir(filepath)
	if err != nil {
		return fmt.Errorf("cannot read the path %s: %s", filepath, err)
	}

	for _, entry := range ls {
		fullpath := path.Join(filepath, entry.Name())

		if entry.IsDir() {
			if v.validDir(entry.Name()) {
				// Scan recursively the directories
				return v.scan(fullpath, ext)
			}
		} else if path.Ext(entry.Name()) == ext {
			// Add the file to the list
			v.results = append(v.results, fullpath)
		}
	}

	return nil
}

// Returns true if the directory name is worth scanning.
func (v *visitor) validDir(name string) bool {
	return name != ".svn" && name != ".hg" && name != ".git"
}

// Scans folder recursively search for files with the ext
// extension and returns the whole list.
func Scan(folder string, ext string) ([]string, error) {
	v := &visitor{[]string{}}
	if err := v.scan(folder, ext); err != nil {
		return nil, err
	}

	return v.results, nil
}
