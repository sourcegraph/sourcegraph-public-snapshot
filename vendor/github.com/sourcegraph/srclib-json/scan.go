package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type ScanCmd struct{}

type srcFileConfig struct{}

var (
	config  = &srcFileConfig{}
	parser  = flags.NewNamedParser("srclib-json", flags.Default)
	scanCmd = ScanCmd{}

	// filePredicates is a list of predicate functions that check to see if we can
	// recognize / process a given JSON file
	filePredicates = []func(s string) bool{}
)

func init() {
	_, err := parser.AddCommand("scan",
		"scan for JSON files",
		"Scan the directory tree rooted at the current directory for JSON Files",
		&scanCmd)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}

func isJSONFile(fileName string) bool {
	return filepath.Ext(fileName) == ".json"
}

var isExcludedDir = map[string]bool{".git": true, ".hg": true, ".srclib-cache": true}

func (c *ScanCmd) Execute(args []string) error {
	if err := json.NewDecoder(os.Stdin).Decode(&config); err != nil {
		return err
	}
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	units, err := scan(cwd)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(units, "", " ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func scan(dir string) ([]*unit.SourceUnit, error) {
	u := unit.SourceUnit{}
	u.Key.Name = filepath.Base(dir)
	u.Key.Type = "json"
	u.Files = []string{}
	units := []*unit.SourceUnit{&u}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if isExcludedDir[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if isJSONFile(path) {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			if includeJSONFile(relPath) {
				u.Files = append(u.Files, filepath.ToSlash(relPath))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return units, nil
}

func includeJSONFile(path string) bool {
	for _, predicate := range filePredicates {
		if predicate(path) {
			return true
		}
	}
	return false
}
