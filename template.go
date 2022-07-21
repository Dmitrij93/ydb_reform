package main

import (
	"io"
	"log"
	"os"
	"strings"
)

func createTargetFile(fileName string) {
	targetFileName := strings.Split(fileName, ".")[0] + "_ydb_reform.go"
	createFile(targetFileName)
	file, err := os.OpenFile(targetFileName, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Error when opening %s file: %s", targetFileName, err)
	}
	imp := strings.Join(parsedTables_.imp, "\n")
	_, err = io.WriteString(file, imp)
	if err != nil {
		log.Fatalf("Error when writing imports in target file: %s", err)
	}
}

func createFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error when creating file: %s", err)
	}
	defer file.Close()
}
