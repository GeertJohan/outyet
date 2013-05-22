package main

import (
	"html/template"
)

type dataOutyet struct {
	Outyet  bool
	Version string
}

var tmplOutyet = template.Must(template.New("outyet").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<title>Is Go {{.Version}} out yet?</title>
	</head>
	<body>
		<center>
			<h2>Is Go {{.Version}} out yet?</h2>
			<h1>
			{{if .Outyet}}
				<a href="` + changeURLBase + `{{.Version}}">YES!</a>
			{{else}}
				No.
			{{end}}
			</h1>
		</center>
	</body>
</html>
`))
