package main

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/sourcegraph/sourcegraph/dev/sg/adr"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

type templateData struct {
	ADRs []adr.ArchitectureDecisionRecord
}

//go:generate sh -c "TZ=Etc/UTC go run ."
func main() {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles(filepath.Join(repoRoot, "dev", "adr-docs", "index.md.tmpl"))
	if err != nil {
		panic(err)
	}

	adrs, err := adr.List(filepath.Join(repoRoot, "doc", "dev", "adr"))
	if err != nil {
		return
	}

	presenter := templateData{
		ADRs: adrs,
	}

	f, err := os.Create(filepath.Join(repoRoot, "doc", "dev", "adr", "index.md"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = tmpl.Execute(f, &presenter)
	if err != nil {
		panic(err)
	}
}
