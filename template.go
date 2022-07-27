package main

import (
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

func createTargetFile(fileName string) {
	targetFileName := strings.Split(fileName, ".")[0] + "_ydbgen.go"
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
{{- .Package }}
`,
		`
import(
	"context"
	"path"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)
`,
		`{{ if .EmptyVariable }}
var writeTx = table.TxControl(
	table.BeginTx(
		table.WithSerializableReadWrite(),
	),
	table.CommitTx(),
)
{{ end }}`,
		`
func (u *{{ .St.NameTable }}) scanValues() []named.Value {
	return []named.Value{
	{{- range $i, $x := .St.Table_ }}
		named.
		{{- if eq $i 0 }}Required
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
		{{- if $x.NullType }}Nullable{{- end }}
		{{- $x.YDBType }}{{ if eq $x.YDBType "Datetime" }}ValueFromTime{{ else }}Value{{ end }}(u.{{ $x.Field }})),
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
func (ur {{ .St.NameTable }}Repo) fields() string {
	return ` + "`" + ` {{ $table := .St.Table_ }}
	{{- range $i, $x := $table  }}
	{{- if last $i $table }}{{ $x.YDBField }}{{- else }}{{ $x.YDBField }}, {{ end }}
	{{- end }} ` + "`" + `
}
`,
		`
func (ur {{ .St.NameTable }}Repo) values() string {
	return ` + "`" + ` ({{ $table := .St.Table_ }}
	{{- range $i, $x := $table  }}
	{{- if last $i $table }}${{ $x.Field }}{{- else }}${{ $x.Field }}, {{ end }}
	{{- end }}) ` + "`" + `
}
`,
		`
func (ur {{ .St.NameTable }}Repo) table(name string) string {
	res := ` + "` {{ lower .St.NameTable }}s `" + `
	if name != "" {
		res += name + ` + "` `" + `
	}
	return res
}
`,
		`
func (ur {{ .St.NameTable }}Repo) findPrimary() string {
	return ` + "` WHERE" + `{{ range $i, $x := .St.Table_  }}
	{{- if $x.YDBPrimary }}
	{{- if eq $i 0}} {{ $x.YDBField }} = ${{ $x.Field }}
	{{- else }} AND {{ $x.YDBField }} = ${{ $x.Field }}
	{{- end }}
	{{- end }}
	{{- end }} ` + "`" + `
}
`,
		`{{ $table := .St }}
{{- $pr := (index $table.Table_ 0) }}
{{- range $i, $x := $table.Table_  }}{{ if and (ne $i 0) $x.YDBPrimary }}
func (ur {{ $table.NameTable }}Repo) findByFirst() string {
	return ` + "` WHERE " + `{{ $pr.YDBField}} = ${{ $pr.Field }} ` + "`" + `
}
{{ break }}
{{- end }}
{{- end -}}
`,
		`{{ $table := .St }}
{{- $pr := (index $table.Table_ 0) }}
{{- range $i, $x := $table.Table_  }}{{ if and (ne $i 0) $x.YDBPrimary }}
func (ur {{ $table.NameTable }}Repo) firstParam(
	{{- low_capitalize $pr.Field }} {{ $pr.Type -}}
) *table.QueryParameters {
	return table.NewQueryParameters(
		table.ValueParam("${{ $pr.Field }}", types.
		{{- $pr.YDBType }}{{ if eq $pr.YDBType "Datetime" }}ValueFromTime{{ else }}Value{{ end -}}
		({{ low_capitalize $pr.Field }})),
	)
}
{{ break }}
{{- end }}
{{- end -}}
`,
		`
func (ur {{ .St.NameTable }}Repo) primaryParams(
	{{- range $i, $x := .St.Table_  }}
	{{- if eq $i 0 }}{{ low_capitalize $x.Field }} {{ $x.Type }}{{ else }}
	{{- if $x.YDBPrimary }}, {{ low_capitalize $x.Field }} {{ $x.Type }}{{ end }}{{ end }}
	{{- end -}}
) *table.QueryParameters {
	return table.NewQueryParameters(
		{{- range $i, $x := .St.Table_  }}
		{{- if $x.YDBPrimary}}
		table.ValueParam("${{ $x.Field }}", types.
		{{- $x.YDBType }}{{ if eq $x.YDBType "Datetime" }}ValueFromTime{{ else }}Value{{ end -}}
		({{ low_capitalize $x.Field }})),
		{{- end }}
		{{- end }}
	)
}
`,
		`
func (ur *{{ .St.NameTable }}Repo) Get(ctx context.Context,
	{{- range $i, $x := .St.Table_  }}
	{{- if eq $i 0 }} {{ low_capitalize $x.Field }} {{ $x.Type }}{{ else }}
	{{- if $x.YDBPrimary }}, {{ low_capitalize $x.Field }} {{ $x.Type }}{{ end }}{{ end }}
	{{- end -}}
) (u *{{ .St.NameTable }}, err error) {
	u = &{{ .St.NameTable }}{}
	query := ur.declarePrimary() + ` + "`SELECT `" + ` + ur.fields() +
		" FROM " + ur.table("") +
		ur.findPrimary()
	var res result.Result
	err = ur.DB.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, table.DefaultTxControl(), query,
			ur.primaryParams(
			{{- range $i, $x := .St.Table_  }}
			{{- if eq $i 0 }}{{ low_capitalize $x.Field }}{{ else }}
			{{- if $x.YDBPrimary }}, {{ low_capitalize $x.Field }}{{ end }}{{ end }}
			{{- end -}}
),
			options.WithCollectStatsModeBasic(),
		)
		return err
	})
	if err != nil {
		return
	}
	defer func() {
		_ = res.Close()
	}()
	for res.NextResultSet(ctx) {
		for res.NextRow() {
			err = res.ScanNamed(u.scanValues()...)
			return
		}
	}
	return
}
`,
		`{{ $table := .St }}
{{- $pr := (index $table.Table_ 0) }}
{{- range $i, $x := $table.Table_  }}{{ if and (ne $i 0) $x.YDBPrimary }}
func (ur *{{ $table.NameTable }}Repo) GetBy{{$pr.Field}}(ctx context.Context, {{ low_capitalize $pr.Field }} {{ $pr.Type -}}
) (ss []*{{ $table.NameTable }}, err error) {
	query := ur.declarePrimary() + ` + "`SELECT `" + ` + ur.fields() +
		" FROM " + ur.table("") +
		ur.findByFirst()
	var res result.Result
	err = ur.DB.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err = s.Execute(ctx, table.DefaultTxControl(), query,
			ur.firstParam({{- low_capitalize $pr.Field }}),
			options.WithCollectStatsModeBasic(),
		)
		return err
	})
	if err != nil {
		return
	}
	defer func() {
		_ = res.Close()
	}()
	for res.NextResultSet(ctx) {
		for res.NextRow() {
			s := &{{ $table.NameTable }}{}
			err = res.ScanNamed(s.scanValues()...)
			if err != nil {
				return
			}
			ss = append(ss, s)
		}
	}
	return
}
{{ break }}
{{- end }}
{{- end -}}
`,
		`
func (ur *{{ .St.NameTable }}Repo) Insert(ctx context.Context, u *{{ .St.NameTable }}) (err error) {
	{{ if .Funcs.BeforeInsert }}u.BeforeInsert()
	{{ end -}}
	query := ur.declare{{ .St.NameTable }}() + ` + "`INSERT INTO `" + ` + ur.table("") + ` + "` (` + ur.fields() + `) VALUES `" + ` + ur.values()
	return ur.DB.Table().Do(
		ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, _, err = s.Execute(ctx, writeTx, query,
				table.NewQueryParameters(u.setValues()...),
				options.WithCollectStatsModeBasic(),
			)
			return err
		},
	)
}
`,
		`
func (ur *{{ .St.NameTable }}Repo) Upsert(ctx context.Context, u *{{ .St.NameTable }}) (err error) {
	{{ if .Funcs.BeforeUpdate }}u.BeforeUpdate()
	{{ end -}}
	query := ur.declare{{ .St.NameTable }}() + ` + "`UPSERT INTO `" + ` + ur.table("") + ` + "` (` + ur.fields() + `) VALUES `" + ` + ur.values()
	return ur.DB.Table().Do(
		ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, _, err = s.Execute(ctx, writeTx, query,
				table.NewQueryParameters(u.setValues()...),
				options.WithCollectStatsModeBasic(),
			)
			return err
		},
	)
}
`,
		`
func (ur *{{ .St.NameTable }}Repo) Delete(ctx context.Context,
	{{- range $i, $x := .St.Table_  }}
	{{- if eq $i 0 }} {{ low_capitalize $x.Field }} {{ $x.Type }}{{ else }}
	{{- if $x.YDBPrimary }}, {{ low_capitalize $x.Field }} {{ $x.Type }}{{ end }}{{ end }}
	{{- end -}}
) (err error) {
	query := ur.declarePrimary() + ` + "`DELETE FROM `" + ` + ur.table("") + ur.findPrimary()
	return ur.DB.Table().Do(
		ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, _, err = s.Execute(ctx, writeTx, query,
				ur.primaryParams(
					{{- range $i, $x := .St.Table_  }}
					{{- if eq $i 0 }}{{ low_capitalize $x.Field }}{{ else }}
					{{- if $x.YDBPrimary }}, {{ low_capitalize $x.Field }}{{ end }}{{ end }}
					{{- end -}}
				),
				options.WithCollectStatsModeBasic(),
			)
			return err
		},
	)
}
`,
		`{{ $table := .St }}
{{- $pr := (index $table.Table_ 0) }}
{{- range $i, $x := $table.Table_  }}{{ if and (ne $i 0) $x.YDBPrimary }}
func (ur *{{ $table.NameTable }}Repo) DeleteBy{{ $pr.Field }}(ctx context.Context, {{ low_capitalize $pr.Field }} {{ $pr.Type }}) (err error) {
	query := ur.declarePrimary() + ` + "`DELETE FROM `" + ` + ur.table("") + ur.findByFirst()
	return ur.DB.Table().Do(
		ctx,
		func(ctx context.Context, s table.Session) (err error) {
			_, _, err = s.Execute(ctx, writeTx, query,
				ur.firstParam({{ low_capitalize $pr.Field }}),
				options.WithCollectStatsModeBasic(),
			)
			return err
		},
	)
}
{{ break }}
{{- end }}
{{- end -}}
`,
		`
func (ur *{{ .St.NameTable }}Repo) CreateTable(ctx context.Context) (err error) {
	return ur.DB.Table().Do(
		ctx,
		func(ctx context.Context, s table.Session) (err error) {
			return s.CreateTable(ctx, path.Join(ur.DB.Name(), "{{ lower .St.NameTable }}s"),
				{{- range $i, $x := .St.Table_  }}
				options.WithColumn("{{ $x.YDBField }}", types.Optional(types.Type
					{{- $x.YDBType }})),
				{{- end }}
				options.WithPrimaryKeyColumn(
					{{- range $i, $x := .St.Table_  }}
					{{- if eq $i 0 }}"{{ $x.YDBField }}"{{ else }}
					{{- if $x.YDBPrimary }}, "{{ $x.YDBField }}"{{ end }}{{ end }}
					{{- end -}}
				),
			)
		},
	)
}
`,
	}
	for i, f := range text {
		t := template.New("").Funcs(
			template.FuncMap{
				"last": func(x int, a interface{}) bool {
					return x == reflect.ValueOf(a).Len()-1
				},
				"low_capitalize": func(s string) string {
					return strings.ToLower(s[:1]) + s[1:]
				},
				"lower": func(s string) string {
					ans := []string{}
					for i, x := range s {
						symb := string(x)
						if i == 0 {
							ans = append(ans, strings.ToLower(symb))
						} else if symb == strings.ToLower(symb) {
							ans = append(ans, symb)
						} else {
							ans = append(ans, "_"+strings.ToLower(symb))
						}

					}
					return strings.Join(ans, "")
				},
			})
		t.Parse(f)
		if err := t.Execute(&file, parsedTables_); err != nil {
			log.Fatalf("Error when writing struct in target file (record %d): %s", i, err)
		}
	}

}
