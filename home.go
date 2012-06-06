package main

import ()

func HomeHandler(r *Request) error {
	return r.ExecuteTemplate(HOME_TEMPLATE, nil)
}

const HOME_TEMPLATE = `
{{define "base"}}
<!DOCTYPE html>
<html>
<head>

	<meta charset="utf-8">
	<title>Home</title>

</head>
<body>

	<h1>Actions</h1>
	<ul>
		<li><a href="/compile">Compiled output</a></li>
		<li><a href="/test/all">MultiTest runner</a></li>
	</ul>

</body>
</html>
{{end}}
`
