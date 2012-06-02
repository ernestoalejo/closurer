package main

import (
	"html/template"
	"io"
)

var templatesCache = map[string]*template.Template{}
var templatesFuncs = template.FuncMap{
	"int_eq": func(a, b int) bool { return a == b },
	"str_eq": func(a, b string) bool { return a == b },
	"bhtml":  func(a []byte) template.HTML { return template.HTML(a) },
	"add":    func(a, b int) int { return a + b },
}

func (r *Request) ExecuteTemplate(names []string, data map[string]interface{}) error {
	// Insert common data
	data = CommonData(r, data)

	// Insert the base template in the list
	names = append(names, "base")

	// Parse the template & execute it
	return RawExecuteTemplate(r.W, names, data)
}

func RawExecuteTemplate(w io.Writer, names []string, data map[string]interface{}) error {
	// Build the key for this template
	cname := ""
	for i, name := range names {
		names[i] = "templates/" + name + ".html"
		cname += name
	}

	t, ok := templatesCache[cname]
	if !ok {
		var err error
		t, err = template.New(cname).Funcs(templatesFuncs).ParseFiles(names...)
		if err != nil {
			return InternalErr(err, "cannot parse the template")
		}
		templatesCache[cname] = t
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return InternalErr(err, "cannot execute the template")
	}

	return nil
}

func CommonData(r *Request, data map[string]interface{}) map[string]interface{} {
	data["R"] = r

	return data
}
