// Package gist6418290 implements ability to get name of caller funcs and parameters.
package gist6418290

import (
	"bytes"
	"fmt"
	"go/ast"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/shurcooL/go/gists/gist5639599"
	"github.com/shurcooL/go/gists/gist5707298"
	"github.com/shurcooL/go/gists/gist6445065"
)

// Gets the parent func as a string.
func GetParentFuncAsString() string {
	// TODO: Replace use of debug.Stack() with direct use of runtime package...
	// TODO: Use runtime.FuncForPC(runtime.Caller()).Name() to get func name if source code not found.
	stack := string(stack())

	funcName := getLine(stack, 3)
	funcName = funcName[1:strings.Index(funcName, ": ")]
	if dotPos := strings.LastIndex(funcName, "."); dotPos != -1 { // Trim package prefix.
		funcName = funcName[dotPos+1:]
	}

	funcArgs := getLine(stack, 5)
	funcArgs = funcArgs[strings.Index(funcArgs, ": ")+len(": "):]
	funcArgs = funcArgs[strings.Index(funcArgs, "(") : strings.LastIndex(funcArgs, ")")+len(")")] // TODO: This may fail if there are 2+ func calls on one line.

	return funcName + funcArgs
}

// Gets the parent func with its args as a string.
func GetParentFuncArgsAsString(args ...interface{}) string {
	// TODO: Replace use of debug.Stack() with direct use of runtime package...
	// TODO: Use runtime.FuncForPC(runtime.Caller()).Name() to get func name if source code not found.
	stack := string(stack())

	funcName := getLine(stack, 3)
	funcName = funcName[1:strings.Index(funcName, ": ")]
	if dotPos := strings.LastIndex(funcName, "."); dotPos != -1 { // Trim package prefix.
		funcName = funcName[dotPos+1:]
	}

	funcArgs := "("
	for i, arg := range args {
		// TODO: Add arg names. Maybe not?
		if i != 0 {
			funcArgs += ", "
		}
		funcArgs += fmt.Sprintf("%#v", arg) // TODO: Maybe use goon instead. Need to move elsewhere to avoid import cycle.
	}
	funcArgs += ")"

	return funcName + funcArgs
}

// Gets the expression as a string.
func GetExprAsString(_ interface{}) string {
	return GetParentArgExprAsString(0)
}

func getParent2ArgExprAllAsAst() []ast.Expr {
	// TODO: Replace use of debug.Stack() with direct use of runtime package...
	stack := string(stack())

	// TODO: Bounds error checking, get rid of GetLine gists, etc.
	parentName := getLine(stack, 5)
	if strings.Index(parentName, ": ") == -1 {
		// TODO: This happens when source file isn't present in same location as when built. See if can do anything better
		//       via direct use of runtime package (instead of debug.Stack(), which will exclude any func names)...
		return nil
	}
	parentName = parentName[1:strings.Index(parentName, ": ")]
	if dotPos := strings.LastIndex(parentName, "."); dotPos != -1 { // Trim package prefix.
		parentName = parentName[dotPos+1:]
	}

	str := getLine(stack, 7)
	str = str[strings.Index(str, ": ")+len(": "):]
	p, err := gist5707298.ParseStmt(str)
	if err != nil {
		return nil
	}

	innerQuery := func(i interface{}) bool {
		if ident, ok := i.(*ast.Ident); ok && ident.Name == parentName {
			return true
		}
		return false
	}

	query := func(i interface{}) bool {
		if c, ok := i.(*ast.CallExpr); ok && nil != gist6445065.FindFirst(c.Fun, innerQuery) {
			return true
		}
		return false
	}
	callExpr, _ := gist6445065.FindFirst(p, query).(*ast.CallExpr)

	if callExpr == nil {
		return nil
	}
	return callExpr.Args
}

// Gets the argIndex argument expression of parent func call as a string.
func GetParentArgExprAsString(argIndex uint32) string {
	args := getParent2ArgExprAllAsAst()
	if args == nil {
		return "<expr not found>"
	}
	if argIndex >= uint32(len(args)) {
		return "<out of range>"
	}

	return gist5639599.SprintAstBare(args[argIndex])
}

// Gets all argument expressions of parent func call as a string.
func GetParentArgExprAllAsString() []string {
	args := getParent2ArgExprAllAsAst()
	if args == nil {
		return nil
	}

	out := make([]string, len(args))
	for i := range args {
		out[i] = gist5639599.SprintAstBare(args[i])
	}
	return out
}

func getMySecondArgExprAsString(int, int) string {
	return GetParentArgExprAsString(1)
}

func getLine(s string, lineIndex int) string {
	return strings.Split(s, "\n")[lineIndex]
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a formatted stack trace of the goroutine that calls it.
// For each routine, it includes the source line information and PC value,
// then attempts to discover, for Go functions, the calling function or
// method and the text of the line containing the invocation.
//
// It was deprecated in Go 1.5, suggested to use package runtime's Stack instead,
// and replaced by another implementation in Go 1.6.
//
// stack implements the Go 1.5 version of debug.Stack(), skipping 1 frame,
// instead of 2, since it's being called directly (rather than via debug.Stack()).
func stack() []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := 1; ; i++ { // Caller we care about is the user, 1 frame up
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		line-- // in stack trace, lines are 1-indexed but our array is 0-indexed
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.Trim(lines[n], " \t")
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	// 	runtime/debug.*T·ptrmethod
	// and want
	// 	*T.ptrmethod
	// Since the package path might contains dots (e.g. code.google.com/...),
	// we first remove the path prefix if there is one.
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
