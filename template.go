package main

import (
	"log"
	"os"
	"reflect"
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
	writeData(*file)
	file.Close()

}

func createFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error when creating file: %s", err)
	}
	defer file.Close()
}

func writeData(file os.File) {

	text := []string{
		`
{{ .Package }}
`,
		`
import(
	"context"
	"path"
	{{- range $i, $x := .Imp }}
	{{ $x }}
	{{- end }}

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)
`,
		`
{{- range $i, $x := .NotParceSt }}"{{ $x }}"
{{ end }}
`,
		`
func (u *{{ .St.NameTable }}) scanValues() []named.Value {
	return []named.Value{
	{{- range $i, $x := .St.Table_ }}
		named.
		{{- if eq $i 0 }}Reqired
		{{- else }}{{if eq $x.NullType false}}OptionalWithDefault
		{{- else }}Optional
		{{- end }}{{ end -}}
		("{{ $x.YDBField }}", &u.{{ $x.Field }}),
		{{- end }}
	}
}
`,
		`
func (u *{{ .St.NameTable }}) setValues() []table.ParameterOption {
	return []table.ParameterOption{
		{{- range $i, $x := .St.Table_ }}
		table.ValueParam("${{ $x.Field }}", types.
		{{- if $x.NullType }}Nullable{{- end -}}
		{{ $x.YDBType }}Value(u.{{ $x.Field }})),
		{{- end }}
	}
}
`,
		`
type {{ .St.NameTable }}Repo struct {
	DB ydb.Connection
}
`,
		`
func (ur {{ .St.NameTable }}Repo) declarePrimary() string {
	return ` + "`" + `
		{{- range $i, $x := .St.Table_  }}
		{{- if $x.YDBPrimary }}
		DECLARE ${{ $x.Field }} AS {{ $x.YDBType }}
		{{- if $x.NullType }}?{{ end }};
		{{- end }}
		{{- end }}
	` + "`" + `
}
`,
		`
func (ur {{ .St.NameTable }}Repo) declare{{ .St.NameTable }}() string {
	return ` + "`" + `
		{{- range $i, $x := .St.Table_  }}
		DECLARE ${{ $x.Field }} AS {{ $x.YDBType }}
		{{- if $x.NullType }}?{{ end }};
		{{- end }}
	` + "`" + `
}
`,
		`
func (ur ProfileRepo) fields() string {
	return ` + "`" + ` {{ $table := .St.Table_ }}
	{{- range $i, $x := $table  }}
	{{- if last $i $table }}{{ $x.YDBField }}{{- else }}{{ $x.YDBField }}, {{ end }}
	{{- end }} ` + "`" + `
}
`,
		`
func (ur ProfileRepo) values() string {
	return ` + "`" + ` ({{ $table := .St.Table_ }}
	{{- range $i, $x := $table  }}
	{{- if last $i $table }}${{ $x.Field }}{{- else }}${{ $x.Field }}, {{ end }}
	{{- end }}) ` + "`" + `
}
`,
	}
	for i, f := range text {
		t := template.New("").Funcs(
			template.FuncMap{
				"last": func(x int, a interface{}) bool {
					return x == reflect.ValueOf(a).Len()-1
				},
			})
		t.Parse(f)
		if err := t.Execute(&file, parsedTables_); err != nil {
			log.Fatalf("Error when writing struct in target file (record %d): %s", i, err)
		}
	}

}
