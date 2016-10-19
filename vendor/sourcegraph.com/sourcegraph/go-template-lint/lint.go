package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"text/template/parse"

	"sourcegraph.com/sourcegraph/go-template-lint/tmplwalk"
)

var (
	verbose      = flag.Bool("v", false, "show verbose output")
	funcmapFile  = flag.String("f", "", "Go source file with FuncMap literal (commas separate multiple files)")
	tmplSetsFile = flag.String("t", "", "Go source file containing template set (a [][]string literal)")
	tmplDir      = flag.String("td", "", "base path of templates (prepended to template set filenames)")

	fset = token.NewFileSet()
)

func main() {
	log.SetFlags(0)

	flag.Parse()
	if *funcmapFile == "" {
		log.Fatal("-f is required (run with -h for usage info)")
	}
	if *tmplSetsFile == "" {
		log.Fatal("-t is required (run with -h for usage info)")
	}

	var allDefinedFuncs []string
	for _, f := range strings.Split(*funcmapFile, ",") {
		definedFuncs, err := parseFuncDefs(f)
		if err != nil {
			log.Fatalf("Error parsing FuncMap names from %s: %s", f, err)
		}
		if len(definedFuncs) == 0 {
			log.Fatalf("No func definitions in a FuncMap found in %s.", f)
		}
		allDefinedFuncs = append(allDefinedFuncs, definedFuncs...)
	}
	if *verbose {
		log.Printf("# Found %d template functions:", len(allDefinedFuncs))
		for _, name := range allDefinedFuncs {
			log.Printf("# - %s", name)
		}
	}

	tmplSets, err := parseTmplSet(*tmplSetsFile)
	if err != nil {
		log.Fatalf("Error parsing template sets (a [][]string literal) from %s: %s", *tmplSetsFile, err)
	}
	if len(tmplSets) == 0 {
		log.Fatal("No template files found.")
	}
	if *verbose {
		log.Printf("# Found %d template file sets:", len(tmplSets))
		for _, ts := range tmplSets {
			log.Printf("# - %s", strings.Join(ts, " "))
		}
	}

	invokedFuncs, err := findInvokedFuncs(tmplSets, allDefinedFuncs)
	if err != nil {
		log.Fatalf("Error parsing templates to find invoked template functions: %s", err)
	}
	if len(invokedFuncs) == 0 {
		log.Fatal("No func invocations in templates found.")
	}
	if *verbose {
		log.Printf("# Found %d functions invoked in templates:", len(invokedFuncs))
		for _, name := range invokedFuncs {
			log.Printf("# - %s", name)
		}
	}

	definedFuncMap := sliceToMap(allDefinedFuncs)
	invokedFuncMap := sliceToMap(invokedFuncs)
	fail := false

	unusedFuncs := subtract(definedFuncMap, invokedFuncMap)
	for _, f := range unusedFuncs {
		fmt.Println("unused template func", f)
		fail = true
	}

	addPredefs(definedFuncMap)
	undefinedFuncs := subtract(invokedFuncMap, definedFuncMap)
	for _, f := range undefinedFuncs {
		fmt.Println("undefined template func", f)
		fail = true
	}

	if fail {
		os.Exit(1)
	}
}

// parseFuncDefs extracts and returns function names from a Go source
// file containing a FuncMap literal.
func parseFuncDefs(filename string) ([]string, error) {
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var funcNames []string
	ast.Walk(visitFn(func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CompositeLit:
			if isFuncMap(n.Type) {
				for _, e := range n.Elts {
					kv := e.(*ast.KeyValueExpr)
					name, err := strconv.Unquote(kv.Key.(*ast.BasicLit).Value)
					if err != nil {
						log.Fatal(err)
					}
					funcNames = append(funcNames, name)
				}
			}
		}
		return true
	}), f)
	return funcNames, nil
}

func isFuncMap(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.Ident:
		return x.Name == "FuncMap"
	case *ast.SelectorExpr:
		return isFuncMap(x.Sel)
	}
	return false
}

func parseTmplSet(filename string) ([][]string, error) {
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var tmplSets [][]string
	ast.Walk(visitFn(func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CompositeLit:
			switch {
			case isTmplSet(n.Type):
				for _, e := range n.Elts {
					tmplSets = append(tmplSets, astStringSlice(e.(*ast.CompositeLit)))
				}
			case isLayoutSet(n.Type):
				tmplSets = append(tmplSets, astStringSlice(n))
			}
		}
		return true
	}), f)
	return tmplSets, nil
}

func isTmplSet(x ast.Expr) bool {
	if sx, ok := x.(*ast.ArrayType); ok && sx.Len == nil {
		if sx2, ok := sx.Elt.(*ast.ArrayType); ok && sx2.Len == nil {
			if t, ok := sx2.Elt.(*ast.Ident); ok && t.Name == "string" {
				return true
			}
		}
	}
	return false
}

func isLayoutSet(x ast.Expr) bool {
	if sx, ok := x.(*ast.ArrayType); ok && sx.Len == nil {
		if t, ok := sx.Elt.(*ast.Ident); ok && t.Name == "string" {
			return true
		}
	}
	return false
}

func astStringSlice(cl *ast.CompositeLit) []string {
	var ss []string
	for _, e := range cl.Elts {
		s, err := strconv.Unquote(e.(*ast.BasicLit).Value)
		if err != nil {
			log.Fatal(err)
		}
		ss = append(ss, s)
	}
	return ss
}

// visitFn is a wrapper for traversing nodes in the AST
type visitFn func(node ast.Node) (descend bool)

func (v visitFn) Visit(node ast.Node) ast.Visitor {
	descend := v(node)
	if descend {
		return v
	}
	return nil
}

// findInvokedFuncs returns a list of all functions (including
// predefined functions) invoked in Go templates in templateDir
// (recursively).
func findInvokedFuncs(tmplSets [][]string, definedFuncs []string) ([]string, error) {
	definedFuncMap := template.FuncMap{}
	for _, f := range definedFuncs {
		definedFuncMap[f] = func() interface{} { return nil }
	}

	invoked := map[string]struct{}{}
	for _, tmplSet := range tmplSets {
		tt := template.New("")
		tt.Funcs(definedFuncMap)
		_, err := tt.ParseFiles(joinTemplateDir(*tmplDir, tmplSet)...)
		if err != nil {
			return nil, fmt.Errorf("template set %v: %s", tmplSet, err)
		}
		for _, t := range tt.Templates() {
			if t.Tree == nil {
				log.Printf("No template root for %v", t.Name())
				continue
			}
			tmplwalk.Inspect(t.Tree.Root, func(n parse.Node) bool {
				switch n := n.(type) {
				case *parse.IdentifierNode:
					invoked[n.Ident] = struct{}{}
				}
				return true
			})
		}
	}

	var invokedList []string
	for f := range invoked {
		invokedList = append(invokedList, f)
	}
	sort.Strings(invokedList)
	return invokedList, nil
}

func joinTemplateDir(base string, files []string) []string {
	result := make([]string, len(files))
	for i := range files {
		result[i] = filepath.Join(base, files[i])
	}
	return result
}

var predefFuncs = []string{"and", "call", "html", "index", "js", "len", "not", "or", "print", "printf", "println", "urlquery", "eq", "ne", "lt", "le", "gt", "ge"}

// addPredefs adds predefined template functions to fm.
func addPredefs(fm map[string]struct{}) {
	for _, f := range predefFuncs {
		fm[f] = struct{}{}
	}
}

// subtract returns a list of keys in a that are not in b.
func subtract(a, b map[string]struct{}) []string {
	var d []string
	for k := range a {
		if _, inB := b[k]; !inB {
			d = append(d, k)
		}
	}
	sort.Strings(d)
	return d
}

func sliceToMap(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}
