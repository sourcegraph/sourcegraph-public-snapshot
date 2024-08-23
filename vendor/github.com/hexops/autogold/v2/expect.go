package autogold

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Value describes a desired value for a Go test, see Expect for more information.
type Value interface {
	// Equal checks if `got` matches the desired test value, invoking t.Fatal otherwise.
	Equal(t *testing.T, got interface{}, opts ...Option)
}

type value struct {
	line  int
	equal func(t *testing.T, got interface{}, opts ...Option)
}

func (v value) Equal(t *testing.T, got interface{}, opts ...Option) {
	t.Helper()
	v.equal(t, got, opts...)
}

var (
	getPackageNameAndPathCacheMu sync.RWMutex
	getPackageNameAndPathCache   = map[[2]string][2]string{}
)

func getPackageNameAndPath(dir, file string) (name, path string, err error) {
	if isBazel() {
		return bazelGetPackageNameAndPath(dir)
	}
	// If it is cached, fetch it from the cache. This prevents us from doing a semi-costly package
	// load for every test that runs, instead requiring we only do it once per _test.go directory.
	getPackageNameAndPathCacheMu.RLock()
	if v, cached := getPackageNameAndPathCache[[2]string{dir, file}]; cached {
		getPackageNameAndPathCacheMu.RUnlock()
		return v[0], v[1], nil
	}
	getPackageNameAndPathCacheMu.RUnlock()

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedFiles, Tests: true}, dir)
	if err != nil {
		return "", "", err

	}

	var testPkg *packages.Package

find:
	for _, pkg := range pkgs {
		for _, goFile := range pkg.GoFiles {
			if filepath.Base(goFile) == filepath.Base(file) {
				testPkg = pkg
				break find
			}
		}
	}

	if testPkg == nil {
		panic(fmt.Errorf("could not find test package for %s in %s", file, dir))
	}

	getPackageNameAndPathCacheMu.Lock()
	getPackageNameAndPathCache[[2]string{dir, file}] = [2]string{testPkg.Name, testPkg.PkgPath}
	getPackageNameAndPathCacheMu.Unlock()
	return testPkg.Name, testPkg.PkgPath, nil
}

// Expect returns an expected Value which can later be checked for equality against a value a test
// produces.
//
// When `-update` is specified, autogold will find and replace in the test file by looking for an
// invocation of `autogold.Expect(...)` at the same line as the callstack indicates for this function
// call, rewriting the `want` value parameter for you.
func Expect(want interface{}) Value {
	_, _, line, _ := runtime.Caller(1)
	return value{
		line: line,
		equal: func(t *testing.T, got interface{}, opts ...Option) {
			t.Helper()
			var (
				profGetPackageNameAndPath time.Duration
				profStringifyExpect       time.Duration
				profStringifyGot          time.Duration
				profDiff                  time.Duration
				profAcquirePathLock       time.Duration
				profReplaceExpect         time.Duration
			)
			writeProfile := func() {
				prof, _ := strconv.ParseBool(os.Getenv("AUTOGOLD_PROFILE"))
				if !prof {
					return
				}
				fmt.Println("autogold: profile:")
				fmt.Println("  getPackageNameAndPath:", profGetPackageNameAndPath)
				fmt.Println("  stringify (want):     ", profStringifyExpect)
				fmt.Println("  stringify (got):      ", profStringifyGot)
				fmt.Println("  diffing   (got):      ", profDiff)
				fmt.Println("  acquire path lock:    ", profAcquirePathLock)
				fmt.Println("  rewrite autogold.Expect:", profReplaceExpect)
			}

			// Identify the root test name ("TestFoo" in "TestFoo/bar")
			testName := t.Name()
			if strings.Contains(testName, "/") {
				split := strings.Split(testName, "/")
				testName = split[0]
			}

			// Find the path to the calling _test.go, relative to where the test is being run.
			var (
				file string
				ok   bool
			)
			for caller := 1; ; caller++ {
				_, file, _, ok = runtime.Caller(caller)
				if !ok || strings.Contains(file, "_test.go") {
					break
				}
			}
			if !ok {
				t.Fatal("runtime.Caller: returned ok=false")
			}
			pwd, err := os.Getwd()
			if err != nil {
				writeProfile()
				t.Fatal(err)
			}

			// Determine the package name and path of the test file, so we can unqualify types in
			// that package.
			start := time.Now()
			pkgName, pkgPath, err := getPackageNameAndPath(pwd, file)
			profGetPackageNameAndPath = time.Since(start)
			if err != nil {
				writeProfile()
				t.Fatalf("loading package: %v", err)
			}
			opts = append(opts, &option{
				forPackagePath: pkgPath,
				forPackageName: pkgName,
			})

			// Check if the test failed or not by diffing the results.
			start = time.Now()
			wantString := stringify(want, opts)
			profStringifyExpect = time.Since(start)
			start = time.Now()
			gotString := stringify(got, opts)
			profStringifyGot = time.Since(start)
			start = time.Now()
			diff := diff(gotString, wantString, opts)
			profDiff = time.Since(start)
			if diff == "" {
				writeProfile()
				return // test passed
			}

			// Update the test file if so desired.
			if update() {
				// Acquire a file-level lock to prevent concurrent mutations to the _test.go file
				// by parallel tests (whether in-process, or not.)
				start = time.Now()
				profAcquirePathLock = time.Since(start)
				if err != nil {
					writeProfile()
					t.Fatal(err)
				}

				// Replace the autogold.Expect(...) call's `want` parameter with the expression for
				// the value we got.
				start = time.Now()
				testPath, err := filepath.Rel(pwd, file)
				if err != nil {
					writeProfile()
					t.Fatal(err)
				}
				_, err = replaceExpect(t, testPath, testName, line, gotString, true)
				profReplaceExpect = time.Since(start)
				if err != nil {
					writeProfile()
					t.Fatal(fmt.Errorf("autogold: %v", err))
				}
			}
			if *failOnUpdate || !update() {
				writeProfile()
				t.Log(fmt.Errorf("mismatch (-want +got):\n%s", colorDiff(diff)))
				t.FailNow()
			}
		},
	}
}

type fileChanges struct {
	before []byte
	now    []byte
}

func (f *fileChanges) remap(oldLineNumber int) int {
	if f.now == nil {
		return oldLineNumber
	}
	// autogold.Expect call ordering is guaranteed to not have changed, so we leverage this to remap
	// lines.
	oldCallNumber := 0
	foundLine := false
	for n, line := range bytes.Split(f.before, []byte("\n")) {
		if bytes.Contains(line, []byte("autogold.Expect(")) {
			oldCallNumber++
			if n == oldLineNumber-1 {
				foundLine = true
				break
			}
		}
	}
	if !foundLine {
		return oldLineNumber
	}

	callNumber := 0
	for n, line := range bytes.Split(f.now, []byte("\n")) {
		if bytes.Contains(line, []byte("autogold.Expect(")) {
			callNumber++
			if callNumber == oldCallNumber {
				return n + 1
			}
		}
	}
	panic("autogold: failed to find new call number; this is a bug please file an issue with a reproducable test case")
}

func (f *fileChanges) update(contents []byte) {
	f.now = contents
}

var changesByFile = map[string]*fileChanges{}

// replaceExpect replaces the invocation of:
//
//	autogold.Expect(...)
//
// With:
//
//	autogold.Expect(<replacement>)
//
// Based on the callstack location of the invocation provided, returning an error if it cannot be
// found.
//
// The returned updated file contents have the specified replacement, with goimports ran over the
// result.
func replaceExpect(t *testing.T, testFilePath, testName string, line int, replacement string, writeFile bool) ([]byte, error) {
	unlock, err := acquirePathLock(testFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := unlock(); err != nil {
			t.Fatal(fmt.Errorf("autogold: %v", err))
		}
	}()

	testFileSrc, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		return nil, err
	}
	changes, ok := changesByFile[testFilePath]
	if !ok {
		changes = &fileChanges{before: testFileSrc}
		changesByFile[testFilePath] = changes
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, testFilePath, testFileSrc, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %v", err)
	}

	// Locate the autogold.Expect() call expression and perform string replacement on its 2nd
	// argument.
	//
	// We use string replacement instead of direct ast.Expr swapping so as to ensure that we
	// can use gofumpt to format just our generated ast.Expr, and just gofmt for the remainder
	// of the file (i.e. leaving final formatting of the file up to the user without us having to
	// provide an option.) For why it is important that we use gofumpt on our generated ast.Expr,
	// see https://github.com/hexops/valast/pull/4. As for "why gofmt(goimports) and not gofumpt
	// on the final file?", simply because gofmt is a superset of gofumpt and we don't want to make
	// the call of using gofumpt on behalf of the user.
	callExpr, err := findExpectCallExpr(fset, f, testName, changes.remap(line))
	if err != nil {
		return nil, err
	}
	arg := callExpr.Args[0]
	start := testFileSrc[:fset.Position(arg.Pos()).Offset]
	end := testFileSrc[fset.Position(arg.End()).Offset:]

	newFile := make([]byte, 0, len(testFileSrc))
	newFile = append(newFile, start...)
	newFile = append(newFile, []byte(replacement)...)
	newFile = append(newFile, end...)
	preFormattingFile := newFile
	newFile, err = imports.Process(testFilePath, newFile, nil)
	if err != nil {
		debug, _ := strconv.ParseBool(os.Getenv("AUTOGOLD_DEBUG"))
		if debug {
			fmt.Println("-------------")
			fmt.Println("ERROR FORMATTING FILE:", err)
			fmt.Println("TEST FILE PATH:", testFilePath)
			fmt.Println("CONTENTS:")
			fmt.Println("-------------")
			fmt.Println(string(preFormattingFile))
			fmt.Println("-------------")
		}
		return nil, fmt.Errorf("formatting file: %v", err)
	}

	changes.update(newFile)

	if writeFile {
		info, err := os.Stat(testFilePath)
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(testFilePath, []byte(newFile), info.Mode()); err != nil {
			return nil, err
		}
	}
	return newFile, nil
}

func findExpectCallExpr(fset *token.FileSet, f *ast.File, testName string, line int) (*ast.CallExpr, error) {
	var foundCallExpr *ast.CallExpr
	pre := func(cursor *astutil.Cursor) bool {
		node := cursor.Node()
		if foundCallExpr != nil {
			return false
		}
		ce, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if !isExpectSelectorExpr(se) {
			return true
		}
		if len(ce.Args) != 1 {
			return true
		}
		position := fset.Position(ce.Args[0].Pos())
		if position.Line != line {
			return true
		}
		foundCallExpr = ce
		return true
	}
	f = astutil.Apply(f, pre, nil).(*ast.File)
	if foundCallExpr == nil {
		return nil, fmt.Errorf("%s: could not find autogold.Expect(…) function call on line %v", fset.File(f.Pos()).Name(), line)
	}
	return foundCallExpr, nil
}

func isExpectSelectorExpr(v *ast.SelectorExpr) bool {
	if v.Sel.Name != "Expect" {
		return false
	}
	ident, ok := v.X.(*ast.Ident)
	if !ok {
		return false
	}
	// TODO: handle renamed import
	return ident.Name == "autogold"
}
