package main

import ()

const TEST_TEMPLATE = `
{{define "base"}}
<DOCTYPE html>
<html>
<head>

	<meta charset="utf-8">
	<title>Unit Test</title>

	<script type="text/javascript" src="/input/base.js"></script>
	<script type="text/javascript" src="/input/{{.Name}}"></script>

</head>
<body>

</body>
</html>
{{end}}
`

func TestHandler(r *Request) error {
	name := r.Req.URL.Path[6:]
	name = name[:len(name)-5] + ".js"

	r.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Name": name,
	}
	return r.ExecuteTemplate(TEST_TEMPLATE, data)
}
