package main

import (
	"path/filepath"

	"github.com/ernestokarim/closurer/app"
	"github.com/ernestokarim/closurer/config"
	"github.com/ernestokarim/closurer/scan"
)

type TestData struct {
	Name string
}

func Test(r *app.Request) error {
	name := r.Req.URL.Path[6:]
	name = name[:len(name)-5] + ".js"

	tdata := &TestData{
		Name: name,
	}
	return r.ExecuteTemplate([]string{"test"}, tdata)
}

type TestListData struct {
	AllTests []string
}

func TestAll(r *app.Request) error {
	tests, err := ScanTests()
	if err != nil {
		return err
	}

	tdata := &TestListData{
		AllTests: tests,
	}
	return r.ExecuteTemplate([]string{"global-test"}, tdata)
}

func TestList(r *app.Request) error {
	tests, err := ScanTests()
	if err != nil {
		return err
	}

	tdata := &TestListData{
		AllTests: tests,
	}
	return r.ExecuteTemplate([]string{"test-list"}, tdata)
}

// Search for "_test.js" files and relativize them to
// the root directory. It replaces the .js ext with .html.
func ScanTests() ([]string, error) {
	conf := config.Current()

	tests, err := scan.Do(conf.RootJs, "_test.js")
	if err != nil {
		return nil, err
	}

	for i, test := range tests {
		// Relativize the path adding .html instead of .js
		p, err := filepath.Rel(conf.RootJs, test[:len(test)-2]+"html")
		if err != nil {
			return nil, app.Error(err)
		}
		tests[i] = p
	}

	return tests, nil
}
