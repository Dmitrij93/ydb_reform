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
type template struct {
	imp []string
	st  map[string]table
}

var Template = template{[]string{}, map[string]table{}}
var imp []string
var currentlyNameTable string = ""
var flag = ""

func parse(text string) {
	if (len(text) < 4 || text[:4] != "type") && flag == "" {
		imp = append(imp, text)
	} else if len(text) > 4 && text[:4] == "type" {
		flag = "parsing struct"
		currentlyNameTable = strings.Fields(text)[1]
		Template.imp = imp
		Template.st[currentlyNameTable] = table{}
	} else if text == "}" {
		flag = "parse is complete"
	} else if flag == "parsing struct" {
		tableField := strings.Fields(text)
		fieldstruct := fieldStruct{tableField[0], tableField[1]}
		ydbFieldString := strings.Split(tableField[2], `"`)[1]
		ydbField := strings.Split(ydbFieldString, ",")
		Template.st[currentlyNameTable][fieldstruct] = ydbField
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
			file, err := os.Open(file.Name())
			if err != nil {
				log.Fatalf("Error when opening file: %s", err)
			}

			fileScanner := bufio.NewScanner(file)
			count := 0
			imp = []string{}
			for fileScanner.Scan() {
				text := fileScanner.Text()
				if count == 0 && text != "//ydb_reform" {
					break
				} else if count != 0 {
					parse(text)
				}
				count++
			}
			//fmt.Println(Template)
			if err := fileScanner.Err(); err != nil {
				log.Fatalf("Error while reading file: %s", err)
			}
			file.Close()
		}

	}
}
