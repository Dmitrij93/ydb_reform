package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type fieldStruct struct{ field, type_ string }
type table map[fieldStruct][]string
type parsedTables struct {
	imp []string
	st  map[string]table
}

var parsedTables_ parsedTables
var imp []string
var currentlyNameTable string = ""
var flag string

func parse(text string) {
	if (len(text) < 4 || text[:4] != "type") && flag == "" {
		imp = append(imp, text)
	} else if len(text) > 4 && text[:4] == "type" {
		flag = "parsing struct"
		currentlyNameTable = strings.Fields(text)[1]
		parsedTables_.imp = imp
		parsedTables_.st[currentlyNameTable] = table{}
	} else if text == "}" {
		flag = "parse is complete"
	} else if flag == "parsing struct" {
		tableField := strings.Fields(text)
		fieldstruct := fieldStruct{tableField[0], tableField[1]}
		ydbFieldString := strings.Split(tableField[2], `"`)[1]
		ydbField := strings.Split(ydbFieldString, ",")
		parsedTables_.st[currentlyNameTable][fieldstruct] = ydbField
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
			parsedTables_ = parsedTables{[]string{}, map[string]table{}}
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
