// Package gist6418462 implements functionality to get a string containing source code of provided func.
package gist6418462

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"runtime"

	. "github.com/shurcooL/go/gists/gist5639599"
	. "github.com/shurcooL/go/gists/gist6433744"
	. "github.com/shurcooL/go/gists/gist6445065"
)

// GetSourceAsString returns the source of the func f.
func GetSourceAsString(f interface{}) string {
	// No need to check for f being nil, since that's handled below.
	fv := reflect.ValueOf(f)
	return GetFuncValueSourceAsString(fv)
}

// GetFuncValueSourceAsString returns the source of the func value fv.
func GetFuncValueSourceAsString(fv reflect.Value) string {
	// Checking the kind catches cases where f was nil, resulting in fv being a zero Value (i.e. invalid kind),
	// as well as when fv is non-func.
	if fv.Kind() != reflect.Func {
		return "kind not func"
	}
	pc := fv.Pointer()
	if pc == 0 {
		return "nil"
	}
	function := runtime.FuncForPC(pc)
	if function == nil {
		return "nil"
	}
	file, line := function.FileLine(pc)

	var startIndex, endIndex int
	{
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return "<file not found>"
		}
		startIndex, endIndex = GetLineStartEndIndicies(b, line-1)
	}

	fs := token.NewFileSet()
	fileAst, err := parser.ParseFile(fs, file, nil, 0*parser.ParseComments)
	if err != nil {
		return "<ParseFile failed>"
	}

	// TODO: Consider using ast.Walk() instead of custom FindFirst()
	query := func(i interface{}) bool {
		// TODO: Factor-out the unusual overlap check
		if f, ok := i.(*ast.FuncLit); ok && ((startIndex <= int(f.Pos())-1 && int(f.Pos())-1 <= endIndex) || (int(f.Pos())-1 <= startIndex && startIndex <= int(f.End())-1)) {
			return true
		}
		return false
	}
	funcAst := FindFirst(fileAst, query)

	// If func literal wasn't found, try again looking for func declaration
	if funcAst == nil {
		query := func(i interface{}) bool {
			// TODO: Factor-out the unusual overlap check
			if f, ok := i.(*ast.FuncDecl); ok && ((startIndex <= int(f.Pos())-1 && int(f.Pos())-1 <= endIndex) || (int(f.Pos())-1 <= startIndex && startIndex <= int(f.End())-1)) {
				return true
			}
			return false
		}
		funcAst = FindFirst(fileAst, query)
	}

	if funcAst == nil {
		return fmt.Sprintf("<func src not found at %v:%v>", file, line)
	}

	return SprintAst(fs, funcAst)
}
