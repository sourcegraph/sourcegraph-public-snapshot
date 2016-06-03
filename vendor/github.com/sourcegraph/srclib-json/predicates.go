package main

import "path/filepath"

//Predicates.go contains predicates that check a given file path
//to see if it is a JSON file that we recognize and support.

func npmPrecicate(path string) bool {
	file := filepath.Base(path)
	return file == "package.json"
}

func typescriptPredicate(path string) bool {
	file := filepath.Base(path)
	_, recognized := map[string]bool{"tsconfig.json": true, "tslint.json": true, "typings.json": true}[file]
	return recognized
}

func meteorPredicate(path string) bool {
	file := filepath.Base(path)
	_, recognized := map[string]bool{"settings.json": true, "versions.json": true}[file]
	return recognized
}

func init() {
	for _, p := range []func(string) bool{npmPrecicate, typescriptPredicate, meteorPredicate} {
		filePredicates = append(filePredicates, p)
	}
}
