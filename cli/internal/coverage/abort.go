package coverage

import (
	"log"
	"strings"
	"sync/atomic"
)

// ErrorCase represents a single concrete error case.
type ErrorCase struct {
	Method, Repo, File, Text string
	Line, Character          int
}

// exceptions is a list of error cases which are ignored.
var exceptions = []ErrorCase{}

// errorContainsExceptions is a list of strings where, when the aborting error
// contains the string, the error is ignored.
var errorContainsExceptions = []string{
	"node is *ast.ArrayType, not ident",
	"node is *ast.AssignStmt, not ident",
	"node is *ast.BasicLit, not ident",
	"node is *ast.BlockStmt, not ident",
	"node is *ast.BinaryExpr, not ident",
	"node is *ast.CompositeLit, not ident",
	"node is *ast.CallExpr, not ident",
	"node is *ast.Field, not ident",
	"node is *ast.FieldList, not ident",
	"node is *ast.File, not ident",
	"node is *ast.GenDecl, not ident",
	"node is *ast.IfStmt, not ident",
	"node is *ast.IndexExpr, not ident",
	"node is *ast.MapType, not ident",
	"node is *ast.ReturnStmt, not ident",
	"node is *ast.RangeStmt, not ident",
	"node is *ast.SelectorExpr, not ident",
	"node is *ast.ForStmt, not ident",
	"node is *ast.UnaryExpr, not ident",
	"node is *ast.ValueSpec, not ident",
	"invalid position line: -",
}

var errCount uint64

// Abort handles exiting the program if the given error is not listed as an
// exception in the above lists.
//
// If debug is true, or fatalCount is greater than zero, the error is logged.
//
// If fatalCount is greater than zero, it controls the error number at which
// Abort will exit the program.
func Abort(debug bool, fatalCount int, err error, e ErrorCase) {
	for _, exception := range errorContainsExceptions {
		if strings.Contains(err.Error(), exception) {
			return
		}
	}
	for _, exception := range exceptions {
		if exception == e {
			return
		}
	}
	if debug || fatalCount > 0 {
		log.Println(err)
	}

	// TODO: atomic
	newCount := atomic.AddUint64(&errCount, 1)
	if fatalCount > 0 && int(newCount) == fatalCount {
		log.Fatalf("%#v\n", e)
	}
}
