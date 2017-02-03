package tools

import (
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/buildutil"
)

// ListPkgsUnderDir is buildutil.ExpandPattern(ctxt, []string{dir +
// "/..."}). The implementation is modified from the upstream
// buildutil.ExpandPattern so we can be much faster. buildutil.ExpandPattern
// looks at all directories under GOPATH if there is a `...` pattern. This
// instead only explores the directories under dir. In future
// buildutil.ExpandPattern may be more performant (there are TODOs for it).
func ListPkgsUnderDir(ctxt *build.Context, dir string) []string {
	ch := make(chan string)

	var wg sync.WaitGroup
	for _, root := range ctxt.SrcDirs() {
		root := root
		wg.Add(1)
		go func() {
			allPackages(ctxt, root, dir, ch)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var pkgs []string
	for p := range ch {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	return pkgs
}

// We use a process-wide counting semaphore to limit
// the number of parallel calls to ReadDir.
var ioLimit = make(chan bool, 20)

// allPackages is from tools/go/buildutil. We don't use the exported method
// since it doesn't allow searching from a directory. We need from a specific
// directory for performance on large GOPATHs.
func allPackages(ctxt *build.Context, root, start string, ch chan<- string) {
	root = filepath.Clean(root) + string(os.PathSeparator)
	start = filepath.Clean(start) + string(os.PathSeparator)

	if strings.HasPrefix(root, start) {
		// If we are a child of start, we can just start at the
		// root. A concrete example of this happening is when
		// root=/goroot/src and start=/goroot
		start = root
	}

	if !strings.HasPrefix(start, root) {
		return
	}

	var wg sync.WaitGroup

	var walkDir func(dir string)
	walkDir = func(dir string) {
		// Avoid .foo, _foo, and testdata directory trees.
		base := filepath.Base(dir)
		if base == "" || base[0] == '.' || base[0] == '_' || base == "testdata" {
			return
		}

		pkg := filepath.ToSlash(strings.TrimPrefix(dir, root))

		// Prune search if we encounter any of these import paths.
		switch pkg {
		case "builtin":
			return
		}

		if pkg != "" {
			ch <- pkg
		}

		ioLimit <- true
		files, _ := buildutil.ReadDir(ctxt, dir)
		<-ioLimit
		for _, fi := range files {
			fi := fi
			if fi.IsDir() {
				wg.Add(1)
				go func() {
					walkDir(filepath.Join(dir, fi.Name()))
					wg.Done()
				}()
			}
		}
	}

	walkDir(start)
	wg.Wait()
}
