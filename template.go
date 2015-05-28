package main

import (
	"html/template"
	"time"
)

type dataOutyet struct {
	Outyet bool
	Number string
}

var tmplOutyet = template.Must(template.New("outyet").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<title>Is Go {{.Number}} out yet?</title>
	</head>
	<body>
		<center>
			<h2>Is Go {{.Number}} out yet?</h2>
			<h1>
			{{if .Outyet}}
				<a href="` + changeURLBase + `{{.Number}}">YES!</a>
			{{else}}
				No.
			{{end}}
			</h1>
			<strong>outyet is currently not working correctly for Go1.5+, see <a href="https://github.com/GeertJohan/outyet/issues/2">github GeertJohan/outyet #2</a>.</strong>
		</center>
	</body>
</html>
`))

type dataStats struct {
	TotalHits   int `bson:"hits"`
	TotalChecks int `bson:"checks"`
	Versions    []*dataStatsVersion
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
							<a href="` + changeURLBase + `{{.Number}}">yes</a>
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
