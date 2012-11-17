package scan

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ernestokarim/closurer/app"
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
			if v.validDir(entry.Name()) {
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
func (v *visitor) validDir(name string) bool {
	return name != ".svn" && name != ".hg" && name != ".git"
}

// Scans folder recursively search for files with the ext
// extension and returns the whole list.
func Do(folder string, ext string) ([]string, error) {
	v := &visitor{[]string{}}
	if err := v.scan(folder, ext); err != nil {
		return nil, err
	}

	return v.results, nil
}
