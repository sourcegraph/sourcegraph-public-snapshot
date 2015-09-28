package main

import (
	"flag"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Run myself!

//go:generate go run remove_protobuf_json_snake_case_tags.go -w _workspace/src/sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph _workspace/src/sourcegraph.com/sourcegraph/go-diff/diff _workspace/src/sourcegraph.com/sourcegraph/go-vcs/vcs _workspace/src/sourcegraph.com/sourcegraph/srclib/graph _workspace/src/sourcegraph.com/sourcegraph/srclib/unit _workspace/src/sourcegraph.com/sourcegraph/vcsstore/vcsclient

// Eliminates the snake_cased JSON field names that protobufs write by
// default. This means we don't have to change the app to use new
// JSON.

var fset = token.NewFileSet()

var (
	overwrite = flag.Bool("w", false, "overwrite files (if false, just prints to stdout)")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	dirs := flag.Args()
	for _, dir := range dirs {
		bpkg, err := build.ImportDir(dir, 0)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range bpkg.GoFiles {
			filename := filepath.Join(bpkg.Dir, file)
			astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
			if err != nil {
				log.Fatal(err)
			}
			processFile(filename, astFile)
		}

	}
}

var jsonTagRegexp = regexp.MustCompile(` ?json:"\w+(,[^"]+)?"`)

func processFile(filename string, f *ast.File) {
	var changed bool
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.TYPE {
			for _, s := range d.Specs {
				ts := s.(*ast.TypeSpec)
				if st, ok := ts.Type.(*ast.StructType); ok {
					for _, field := range st.Fields.List {
						if field.Tag != nil {
							// Remove json struct tag
							newTag := jsonTagRegexp.ReplaceAllString(field.Tag.Value, ` json:"$1"`)
							const emptyJSONTag = `json:""`
							if !strings.Contains(field.Tag.Value, emptyJSONTag) && strings.Contains(newTag, emptyJSONTag) {
								newTag = strings.Replace(newTag, emptyJSONTag, "", 1)
							}
							if newTag != field.Tag.Value {
								field.Tag.Value = newTag
								changed = true
							}
						}
					}
				}
			}
		}
	}

	if changed {
		var w io.Writer
		if *overwrite {
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			w = f
		} else {
			w = os.Stdout
		}
		log.Printf("# %s", filename)
		if err := format.Node(w, fset, f); err != nil {
			log.Fatal(err)
		}
	}
}
