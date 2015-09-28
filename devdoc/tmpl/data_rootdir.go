// +build dev

package tmpl

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
)

// rootDir is the directory containing the html/template template files.
var rootDir = filepath.Join(defaultBase("src.sourcegraph.com/sourcegraph/devdoc"), "tmpl")

func defaultBase(path string) string {
	p, err := build.Default.Import(path, "", build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	p.Dir, err = filepath.Rel(cwd, p.Dir)
	if err != nil {
		log.Fatal(err)
	}
	return p.Dir
}
