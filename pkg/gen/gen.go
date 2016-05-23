package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

// Service contains information about a service which has been generated from
// protobuf files. It is used as data for the generator template.
type Service struct {
	Name     string
	PkgName  string
	TypeName string
	Methods  []*Method
}

// Method contains information about a service's method.
type Method struct {
	Name       string
	Type       *ast.FuncType
	ParamType  string
	ResultType string
}

func typeName(e ast.Expr, pkg string, thisPkg string) string {
	var typ string
	switch x := e.(type) {
	case *ast.Ident:
		typ = x.Name

	case *ast.SelectorExpr:
		pkg = x.X.(*ast.Ident).Name
		typ = x.Sel.Name

	case *ast.StarExpr:
		return typeName(x.X, pkg, thisPkg)
	}
	if typ == "" {
		panic("unexpected type")
	}
	if pkg == "pbtypes1" {
		pkg = "pbtypes"
	}
	if pkg == thisPkg {
		return typ
	}
	return fmt.Sprintf("%s.%s", pkg, typ)
}

type serviceList []*Service

func (v serviceList) Len() int           { return len(v) }
func (v serviceList) Less(i, j int) bool { return v[i].TypeName < v[j].TypeName }
func (v serviceList) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

var ifacePat = regexp.MustCompile(`^\w+Server$`)

// Generate takes a list of Go file names and extracts all the interfaces with a
// "Server" suffix. It uses them to render tmpl into the file outFile.
// Optionally, a filter function can be provided to only render a subset of
// services.
func Generate(outFile string, tmpl *template.Template, files []string, filter func(*Service) bool, thisPkg string) {
	fset := token.NewFileSet()
	var svcs []*Service
	for _, file := range files {
		astFile, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
		if err != nil {
			log.Fatal(err)
		}

		ifaces := Types(astFile, func(tspec *ast.TypeSpec) bool {
			_, ok := tspec.Type.(*ast.InterfaceType)
			return ok && ifacePat.MatchString(tspec.Name.Name)
		})
		if len(ifaces) == 0 {
			log.Printf("warning: file %s has no interface types matching %s", file, ifacePat.String())
			continue
		}

		for _, iface := range ifaces {
			if strings.Contains(iface.Name.Name, "_") {
				// Skip Abc_XyzServer interfaces for streaming
				// methods.
				continue
			}
			l := iface.Type.(*ast.InterfaceType).Methods.List
			methods := make([]*Method, 0, len(l))
			for _, m := range l {
				t := m.Type.(*ast.FuncType)

				// Skip streaming methods.
				if arg0Type := typeName(t.Params.List[0].Type, astFile.Name.Name, thisPkg); arg0Type != "context.Context" {
					continue
				}

				methods = append(methods, &Method{
					Name:       m.Names[0].Name,
					Type:       t,
					ParamType:  typeName(t.Params.List[1].Type, astFile.Name.Name, thisPkg),
					ResultType: typeName(t.Results.List[0].Type, astFile.Name.Name, thisPkg),
				})
			}

			svc := &Service{
				Name:     strings.TrimSuffix(iface.Name.Name, "Server"),
				PkgName:  astFile.Name.Name,
				TypeName: astFile.Name.Name + "." + iface.Name.Name,
				Methods:  methods,
			}
			if filter == nil || filter(svc) {
				svcs = append(svcs, svc)
			}
		}
	}

	// Sort for determinism.
	sort.Sort(serviceList(svcs))

	var w bytes.Buffer
	if err := tmpl.Execute(&w, svcs); err != nil {
		log.Fatal(err)
	}

	src, err := format.Source(w.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(src); err != nil {
		log.Fatal(err)
	}
}

func astString(fset *token.FileSet, x ast.Expr) string {
	if x == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
