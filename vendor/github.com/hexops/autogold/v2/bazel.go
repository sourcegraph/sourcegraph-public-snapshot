package autogold

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Bazel build systems try to keep hermeticity by setting PATH="." - but Go does not like this as
// it is a security concern; almost all Go tooling relies on golang.org/x/tools/go/packages.Load
// which behind the scenes must invoke `go` and uses the secure version of x/sys/execabs, but
// ultimately this means Go tools like autogold cannot be run in Bazel:
//
// https://github.com/golang/go/issues/57304
//
// Autogold relies on `packages.Load` in order to determine the Go package name / path when writing
// out a Go AST representation of the value passed in; but the issue above means autogold cannot be
// used with Bazel without removing "." from your PATH, which Bazel claims breaks hermeticity (one
// of the whole reasons people use Bazel.)
//
// For Bazel users, we allow them to set ENABLE_BAZEL_PACKAGES_LOAD_HACK=true which causes autogold
// to guess/infer package names and paths using stack trace information and import paths. This is
// not perfect, it doesn't respect packages whose import paths donot match their defined
// `package foo` statement for example - but it's sufficient to enable autogold to be used in Bazel
// build environments where the above Go/Bazel bug is found.

func isBazel() bool {
	hacks, _ := strconv.ParseBool(os.Getenv("ENABLE_BAZEL_PACKAGES_LOAD_HACK"))
	return hacks
}

// Guesses a package name and import path using Go debug stack trace information.
//
// It looks at the current goroutine's stack, finds the most recent function call in a `_test.go`
// file, and then guesses the package name and path based on the function name.
//
// This does not respect packages whose import path does not match their defined `package autogold_test`
// statement.
//
// This does not respect packages
func bazelGetPackageNameAndPath(dir string) (name, path string, err error) {
	// Guesses an import path based on a function name like:
	//
	// github.com/hexops/autogold/v2.getPackageNameAndPath
	// github.com/hexops/autogold/v2.Expect.func1
	//
	guessPkgPathFromFuncName := func(funcName string) string {
		components := strings.Split(funcName, ".")
		pkgPath := []string{}
		for _, comp := range components {
			pkgPath = append(pkgPath, comp)
			if strings.Contains(comp, "/") {
				break
			}
		}
		return strings.Join(pkgPath, ".")
	}

	var (
		file string
		ok   bool
		pc   uintptr
	)
	for caller := 1; caller < 10000; caller++ {
		pc, file, _, ok = runtime.Caller(caller)
		if !ok {
			break
		}
		if !strings.Contains(file, "_test.go") {
			continue
		}
		pkgPath := guessPkgPathFromFuncName(runtime.FuncForPC(pc).Name())
		pkgName, _ := bazelPackagePathToName(pkgPath)
		return pkgName, pkgPath, nil
	}
	return "", "", errors.New("unable to guess package name/path due to BAZEL_BAD=true")
}

// Guesses a Go package name based on the last component of a Go package path. e.g.:
//
// github.com/hexops/autogold/v2 -> autogold
// github.com/hexops/autogold -> autogold
// cmd/blobstore/internal/blobstore/blobstore_test_test -> blobstore_test
//
// Note that in the third case, Bazel appears to do some reckless renaming of Go package paths,
// where that package would otherwise have path "github.com/sourcegraph/sourcegraph/cmd/blobstore/internal/blobstore"
// and "package blobstore_test" as its name.
//
// This does not respect packages whose import path does not match their defined `package autogold_test`
// statement.
func bazelPackagePathToName(path string) (string, error) {
	components := strings.Split(path, "/")
	last := components[len(components)-1]
	if !strings.Contains(path, ".") {
		// Third case.
		return strings.TrimSuffix(last, "_test"), nil
	}
	if strings.HasPrefix(last, "v") {
		if _, err := strconv.ParseUint(last[1:], 10, 32); err == nil {
			// Package path has a version suffix, e.g. github.com/hexops/autogold/v2
			// and we want the "autogold" component not "v2"
			last = components[len(components)-2]
		}
	}
	return last, nil
}
