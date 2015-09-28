package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	rootDir  = flag.String("rootdir", ".", "root source dir")
	rootPkg  = flag.String("rootpkg", "src.sourcegraph.com/sourcegraph", "root package path (imports not beginning with this path must be found in Godeps/_workspace or GOROOT)")
	useGodep = flag.Bool("godep", true, "prepend {root}/Godeps/_workspace to GOPATH")

	godepGOPATH string

	exitStatus int
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	flag.Parse()

	absRoot, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatal(err)
	}
	godepGOPATH = filepath.Join(absRoot, "Godeps", "_workspace")
	build.Default.GOPATH = godepGOPATH + ":" + build.Default.GOPATH

	var wg sync.WaitGroup
	err = filepath.Walk(*rootDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() && strings.HasPrefix(fi.Name(), "_") {
			return filepath.SkipDir
		}
		if filepath.Ext(path) == ".go" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				checkFile(path)
			}()
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
	os.Exit(exitStatus)
}

func checkFile(path string) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		log.Fatal(err)
	}

	for _, imp := range f.Imports {
		pkg, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			log.Fatal(err)
		}
		checkImport(pkg, path)
	}
}

var (
	// imports holds the import paths of imported packages we have
	// seen and checked.
	imports   = map[string]struct{}{}
	importsMu sync.Mutex
)

func checkImport(path, fromFile string) {
	importsMu.Lock()
	if _, seen := imports[path]; seen {
		importsMu.Unlock()
		return
	}
	imports[path] = struct{}{}
	importsMu.Unlock()

	pkg, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		if strings.HasPrefix(err.Error(), "cannot find package") {
			fmt.Printf("not found:\t%s (from %s)\n", path, fromFile)
			exitStatus = 1
			return
		}
		log.Fatal(err)
	}

	if pathHasPrefix(path, *rootPkg) {
		// pkg is a
		// src.sourcegraph.com/sourcegraph/... package; we
		// assume this is properly added the repo.
		return
	}

	// Pkg must come from Godeps/_workspace or GOROOT.
	if !pkg.Goroot && !pathHasPrefix(pkg.Dir, godepGOPATH) {
		fmt.Printf("not in Godeps:\t%s (from %s)\n", path, fromFile)
		exitStatus = 1
	}
}

func pathHasPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
