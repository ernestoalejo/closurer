package scan

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
)

var (
	libraryCache = map[string][]string{}
)

type visitor struct {
	results []string
}

func (v *visitor) scan(file string, ext string) error {
	ls, err := ioutil.ReadDir(file)
	if err != nil {
		return app.Error(err)
	}

	for _, entry := range ls {
		fullpath := filepath.Join(file, entry.Name())

		if entry.IsDir() {
			if v.validDir(fullpath, entry.Name()) {
				if err := v.scan(fullpath, ext); err != nil {
					return err
				}
			}
		} else if strings.HasSuffix(entry.Name(), ext) {
			v.results = append(v.results, fullpath)
		}
	}

	return nil
}

// Returns true if the directory name is worth scanning.
// It checks too the list of ignored files.
func (v *visitor) validDir(path, name string) bool {
	conf := config.Current()
	if conf.Ignores != nil && path != "" {
		for _, ignore := range conf.Ignores {
			if strings.HasPrefix(path, ignore.Path) {
				return false
			}
		}
	}

	return name != ".svn" && name != ".hg" && name != ".git"
}

// Scans folder recursively search for files with the ext
// extension and returns the whole list.
func Do(folder string, ext string) ([]string, error) {
	conf := config.Current()

	var library bool
	if conf.Library != nil {
		library = strings.Contains(folder, conf.Library.Root)
	}

	if library {
		r, ok := libraryCache[folder]
		if ok {
			return r, nil
		}
	}

	v := &visitor{[]string{}}
	if err := v.scan(folder, ext); err != nil {
		return nil, err
	}

	if library {
		libraryCache[folder] = v.results
	}

	return v.results, nil
}
