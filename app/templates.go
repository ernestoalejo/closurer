package app

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ernestokarim/closurer/cache"
)

const (
	PACKAGE         = "github.com/ernestokarim/closurer"
	DATETIME_FORMAT = "02/01/2006 15:04:05"
)

var (
	templatesCache = map[string]*template.Template{}
	templatesFuncs = template.FuncMap{
		"equals":   func(a, b interface{}) bool { return a == b },
		"last":     func(max, i int) bool { return i == max-1 },
		"datetime": func(t time.Time) string { return t.Format(DATETIME_FORMAT) },
		"bhtml":    func(s string) template.HTML { return template.HTML(s) },
		"nl2br":    func(s string) template.HTML { return template.HTML(strings.Replace(s, "\n", "<br>", -1)) },
	}
)

func RawExecuteTemplate(w io.Writer, names []string, data interface{}) error {
	// Build the key for this template
	cname := ""
	for i, name := range names {
		names[i] = filepath.Join(getPackagePath(), "templates", name+".html")
		cname += name
	}

	// Parse the templates
	t, ok := templatesCache[cname]
	if !ok || cache.NoCache {
		var err error
		t, err = template.New(cname).Funcs(templatesFuncs).ParseFiles(names...)
		if err != nil {
			return fmt.Errorf("cannot parse the template: %s", err)
		}
		templatesCache[cname] = t
	}

	// Execute them
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("cannot execute the template: %s", err)
	}

	return nil
}

func AddTemplateFunc(name string, f interface{}) {
	templatesFuncs[name] = f
}

func getPackagePath() string {
	req := filepath.FromSlash(path.Clean(PACKAGE))
	plist := strings.Split(os.Getenv("GOPATH"), ":")
	for _, p := range plist {
		abs := filepath.Join(p, "src", req)
		if _, err := os.Stat(abs); err != nil && !os.IsNotExist(err) {
			panic(err)
		} else if err == nil {
			return abs
		}
	}

	return ""
}
