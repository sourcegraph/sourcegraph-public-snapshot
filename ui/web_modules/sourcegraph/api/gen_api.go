// +build ignore

package main

import (
	"go/types"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/loader"
)

type apiType struct {
	Name   string
	Fields []*apiField
}

type apiField struct {
	Name     string
	Optional bool
	Type     string
}

func main() {
	var conf loader.Config
	conf.Import("sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph")
	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	packages := []string{
		"sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sourcegraph/srclib/graph",
		"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs",
		"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph",
	}
	apiTypes := make(map[string]*apiType)
	for _, pkg := range packages {
		s := prog.Package(pkg).Pkg.Scope()
		for _, name := range s.Names() {
			o := s.Lookup(name)
			if !o.Exported() || strings.Contains(name, "_") {
				continue
			}
			if s, ok := o.Type().Underlying().(*types.Struct); ok {
				var fields []*apiField
				var addFields func(s *types.Struct)
				addFields = func(s *types.Struct) {
					for i := 0; i < s.NumFields(); i++ {
						f := s.Field(i)
						tag := reflect.StructTag(s.Tag(i)).Get("json")
						optional := strings.Contains(tag, "omitempty")
						if idx := strings.Index(tag, ","); idx != -1 {
							tag = tag[:idx]
						}
						if tag == "" {
							if f.Anonymous() {
								t := f.Type().Underlying()
								if s2, ok := t.(*types.Struct); ok {
									addFields(s2)
								}
							}
							continue
						}
						fields = append(fields, &apiField{Name: tag, Optional: optional, Type: tsType(f.Type())})
					}
				}
				addFields(s)
				apiTypes[name] = &apiType{Name: name, Fields: fields}
			}
		}
	}

	out, err := os.Create("index.tsx")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if err := t.Execute(out, apiTypes); err != nil {
		panic(err)
	}
}

func tsType(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return "boolean"
		case types.String:
			return "string"
		}
		if t.Info()&types.IsNumeric != 0 {
			return "number"
		}
	case *types.Pointer:
		return tsType(t.Elem())
	case *types.Slice:
		return tsType(t.Elem()) + "[]"
	case *types.Named:
		if _, ok := t.Underlying().(*types.Struct); ok && t.Obj().Pkg().Path() == "sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph" {
			return t.Obj().Name()
		}
	}
	return "any"
}

var t = template.Must(template.New("").Parse(`// GENERATED CODE - DO NOT EDIT!
{{range .}}
export interface {{.Name}} {
{{- range .Fields}}
	{{.Name}}{{if .Optional}}?{{end}}: {{.Type}};
{{- end}}
}
{{end}}`))
