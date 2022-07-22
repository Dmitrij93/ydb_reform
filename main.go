package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type table struct {
	Field    string
	Type     []string
	YDBField []string
	YDBType  string
}
type namedTable struct {
	NameTable string
	Table_    []table
}
type parsedTables struct {
	Imp []string
	St  namedTable
}

var parsedTables_ parsedTables
var imp []string
var currentlyNameTable string = ""
var flag string
var ydbType = map[string]string{
	"string":     "UTF8",
	"bool":       "Bool",
	"uint":       "Uint",
	"uint8":      "Uint8",
	"uint16":     "Uint16",
	"uint32":     "Uint32",
	"uint64":     "Uint64",
	"int":        "Int",
	"int8":       "Int8",
	"int16":      "Int16",
	"int32":      "Int32",
	"int64":      "Int64",
	"float32":    "Float",
	"float64":    "Float",
	"complex64":  "",
	"complex128": "",
	"time.Time":  "DatetimeValueFromTime",
}

func parse(text string) {
	if (len(text) < 4 || text[:4] != "type") && flag == "" {
		imp = append(imp, text)
	} else if len(text) > 4 && text[:4] == "type" {
		flag = "parsing struct"
		currentlyNameTable = strings.Fields(text)[1]
		parsedTables_.Imp = imp
		parsedTables_.St = namedTable{currentlyNameTable, []table{}}
	} else if text == "}" {
		flag = "parse is complete"
	} else if flag == "parsing struct" {
		tableField := strings.Fields(text)
		typeField := strings.Split(tableField[1], `*`)
		ydbField := strings.Split(strings.Split(tableField[2], `"`)[1], ",")
		fieldstruct := table{tableField[0], typeField, ydbField, ydbType[typeField[len(typeField)-1]]}
		parsedTables_.St.Table_ = append(parsedTables_.St.Table_, fieldstruct)
	}
}

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatalf("Error when opening file: %s", err)
			}
			fileScanner := bufio.NewScanner(file)
			count := 0
			flag = ""
			imp = []string{}
			parsedTables_ = parsedTables{[]string{}, namedTable{}}
			for fileScanner.Scan() {
				text := fileScanner.Text()
				if count == 0 && text != "//ydb_reform" {
					flag = "not target file"
					break
				} else if count != 0 {
					parse(text)
				}
				count++
			}
			if err := fileScanner.Err(); err != nil {
				log.Fatalf("Error while reading file: %s", err)
			}
			file.Close()
			if flag == "not target file" {
				continue
			}
			createTargetFile(fileName)
		}

	}
}
