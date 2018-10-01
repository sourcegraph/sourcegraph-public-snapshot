// Package goon is a deep pretty printer with Go-like notation. It implements the goon specification.
package goon

import (
	"io"
	"os"

	"github.com/shurcooL/go/reflectsource"
)

// Dump dumps goons to stdout.
func Dump(a ...interface{}) (n int, err error) {
	return os.Stdout.Write(bdump(a...))
}

// Sdump dumps goons to a string.
func Sdump(a ...interface{}) string {
	return string(bdump(a...))
}

// Fdump dumps goons to a writer.
func Fdump(w io.Writer, a ...interface{}) (n int, err error) {
	return w.Write(bdump(a...))
}

// DumpExpr dumps goon expressions to stdout.
//
// E.g., this:
//
// 	somethingImportant := 5
// 	DumpExpr(somethingImportant)
//
// Will print:
//
// 	somethingImportant = (int)(5)
func DumpExpr(a ...interface{}) (n int, err error) {
	return os.Stdout.Write(bdumpNamed(reflectsource.GetParentArgExprAllAsString(), a...))
}

// SdumpExpr dumps goon expressions to a string.
func SdumpExpr(a ...interface{}) string {
	return string(bdumpNamed(reflectsource.GetParentArgExprAllAsString(), a...))
}

// FdumpExpr dumps goon expressions to a writer.
func FdumpExpr(w io.Writer, a ...interface{}) (n int, err error) {
	names := reflectsource.GetParentArgExprAllAsString()
	if len(names) >= 1 {
		names = names[1:] // First argument is the writer, skip it.
	}
	return w.Write(bdumpNamed(names, a...))
}
