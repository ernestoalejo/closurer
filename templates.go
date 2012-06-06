package main

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
)

var (
	globalCname    = -1
	templatesCache = map[string]*template.Template{}
	templatesFuncs = template.FuncMap{
		"int_eq": func(a, b int) bool { return a == b },
		"str_eq": func(a, b string) bool { return a == b },
		"bhtml":  func(a []byte) template.HTML { return template.HTML(a) },
		"add":    func(a, b int) int { return a + b },
	}
)

func (r *Request) ExecuteTemplate(content string, data map[string]interface{}) error {
	// Insert common data
	data = CommonData(r, data)

	// Parse the template & execute it
	return RawExecuteTemplate(r.W, content, data)
}

func RawExecuteTemplate(w io.Writer, content string, data map[string]interface{}) error {
	globalCname += 1
	cname := strconv.Itoa(globalCname)

	// Build the key for this template
	t, ok := templatesCache[cname]
	if !ok {
		var err error
		t, err = template.New(cname).Funcs(templatesFuncs).Parse(content)
		if err != nil {
			return fmt.Errorf("cannot parse the template: %s", err)
		}
		templatesCache[cname] = t
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("cannot execute the template: %s", err)
	}

	return nil
}

func CommonData(r *Request, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}

	data["R"] = r

	return data
}
