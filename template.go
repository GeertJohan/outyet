package main

import (
	"html/template"
	"time"
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

type dataStats struct {
	TotalHits   int `bson:"hits"`
	TotalChecks int `bson:"checks"`
	Versions    []dataStatsVersion
}

type dataStatsVersion struct {
	Number     string    `bson:"number"`
	HitCount   int       `bson:"hits"`
	CheckCount int       `bson:"checks"`
	CreateTime time.Time `bson:"createTime"`
	Outyet     bool
}

var tmplStats = template.Must(template.New("outyet").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<title>Outyet stats</title>
	</head>
	<body>
		Total hits: {{.TotalHits}}<br/>
		Total checks: {{.TotalChecks}}<br/>

		<table>
			<tr>
				<td>Number</td>
				<td>Outyet</td>
				<td>Hits</td>
				<td>Checks</td>
				<td>Created</td>
			</tr>
			{{range .Versions}}
				<tr>
					<td>{{.Number}}</td>
					<td>
						{{if .Outyet}}
							<a href="` + changeURLBase + `{{.Version}}">yes</a>
						{{else}}
							no
						{{end}}
					</td>
					<td>{{.HitCount}}</td>
					<td>{{.CheckCount}}</td>
					<td>{{.CreateTime}}</td>
				</tr>
			{{end}}
		</table>
	</body>
</html>
`))
