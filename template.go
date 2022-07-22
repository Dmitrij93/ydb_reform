package main

import (
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

func createTargetFile(fileName string) {
	targetFileName := strings.Split(fileName, ".")[0] + "_ydb_reform.go"
	createFile(targetFileName)
	file, err := os.OpenFile(targetFileName, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Error when opening %s file: %s", targetFileName, err)
	}
	text := strings.Join(parsedTables_.Imp, "\n")
	_, err = io.WriteString(file, text)
	if err != nil {
		log.Fatalf("Error when writing imports in target file: %s", err)
	}
	appendFuncs(*file)
	file.Close()

}

func createFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error when creating file: %s", err)
	}
	defer file.Close()
}

func appendFuncs(file os.File) {

	text := []string{
		`
func (u *{{ .St.NameTable }}) scanValues() []named.Value {
	return []named.Value{
	{{- range $i, $x := .St.Table_ }}
		named.
		{{- if eq $i 0 }}Reqired
		{{- else }}{{if index $x.Type 0}}OptionalWithDefault
		{{- else }}Optional
		{{- end }}{{ end -}}
		("{{ index $x.YDBField 0 }}", &u.{{ $x.Field }}),
		{{- end }}
	}
}
`,
		`
func (u *{{ .St.NameTable }}) setValues() []table.ParameterOption {
	return []table.ParameterOption{
		{{- range $i, $x := .St.Table_ }}
		table.ValueParam("${{ $x.Field }}", types.
		{{- if index $x.Type 0 }}
		{{- else }}Nullable
		{{- end }}{{ $x.YDBType }}Value(u.{{ $x.Field }})),
		{{- end }}

		table.ValueParam("$UserID", types.Uint64Value(u.UserID)),
		table.ValueParam("$Username", types.UTF8Value(u.Username)),
		table.ValueParam("$FirstName", types.UTF8Value(u.FirstName)),
		table.ValueParam("$LastName", types.UTF8Value(u.LastName)),
		table.ValueParam("$LanguageCode", types.UTF8Value(u.LanguageCode)),
	}
}
`,
		`
type {{ .St.NameTable }}Repo struct {
	DB ydb.Connection
}
`,
	}
	for i, f := range text {
		t := template.New("")
		t.Parse(f)
		if err := t.Execute(&file, parsedTables_); err != nil {
			log.Fatalf("Error when writing struct in target file (record %d): %s", i, err)
		}
	}

}
