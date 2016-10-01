package golang

import (
	"context"
	"fmt"
	"go/build"
	"net/http"
	"path"
	"runtime"
	"sort"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/godoc/vfs"
)

// fetchTransitiveDepsOfFile fetches the transitive dependencies of
// the named Go file. A Go file's dependencies are the imports of its
// own package, plus all of its imports' imports, and so on.
//
// It adds fetched dependencies to its own file system overlay, and
// the returned depFiles should be passed onto the language server to
// add to its overlay.
func (h *BuildHandler) fetchTransitiveDepsOfFile(ctx context.Context, fileURI string) error {
	bctx := h.overlayBuildContext(&build.Context{
		GOOS:     goos,
		GOARCH:   goarch,
		GOPATH:   gopath,
		GOROOT:   goroot,
		Compiler: gocompiler,
	}, false)

	bpkg, err := buildutil.ContainingPackage(bctx, "", h.filePath(fileURI))
	if err != nil && !isMultiplePackageError(err) {
		return err
	}
	// Need to re-import the package because ContainingPackage uses
	// build.FindOnly, which doesn't set the Imports fields.
	bpkg, err = bctx.Import(bpkg.ImportPath, h.rootFSPath, 0)
	if err != nil && !isMultiplePackageError(err) {
		return err
	}

	// Separate mutexes for each VFS source URL.
	var urlMusMu sync.Mutex
	urlMus := map[string]*sync.Mutex{}
	urlMu := func(path string) *sync.Mutex {
		urlMusMu.Lock()
		mu, ok := urlMus[path]
		if !ok {
			mu = new(sync.Mutex)
			urlMus[path] = mu
		}
		urlMusMu.Unlock()
		return mu
	}

	err = doDeps(bpkg, 0, func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		// If the package exists in the repo, or is vendored, or has
		// already been fetched, this will succeed.
		pkg, err := bctx.Import(path, srcDir, mode)
		if isMultiplePackageError(err) {
			err = nil
		}
		if err == nil {
			return pkg, nil
		}

		// Otherwise, it's an external dependency. Fetch the package
		// and try again.
		d, err := resolveImportPath(http.DefaultClient, path)
		if err != nil {
			return nil, err
		}
		if d.vcs != "git" {
			return nil, fmt.Errorf("Go dependency at import path %q has unsupported VCS %q (clone URL is %q)", path, d.vcs, d.cloneURL)
		}
		urlMu := urlMu(d.cloneURL)
		urlMu.Lock()
		defer urlMu.Unlock()

		// Check again after waiting.
		pkg, err = bctx.Import(path, srcDir, mode)
		if err == nil {
			return pkg, nil
		}

		// If not, we hold the lock and we will fetch the dep.
		if err := h.fetchDep(ctx, d); err != nil {
			return nil, err
		}

		pkg, err = bctx.Import(path, srcDir, mode)
		if isMultiplePackageError(err) {
			err = nil
		}
		return pkg, err
	})
	return err
}

func (h *BuildHandler) fetchDep(ctx context.Context, d *directory) error {
	rev := d.rev
	if rev == "" {
		rev = "HEAD"
	}

	fs, err := xlang.CreateGitVFS(fmt.Sprintf("%s?%s#%s", d.cloneURL, rev, ""))
	if err != nil {
		return err
	}

	if _, isStdlib := stdlibPackagePaths[d.importPath]; isStdlib {
		// The zversion.go file is generated during the Go release
		// process and does not exist in the VCS repo archive zips. We
		// need to create it here, or else we'll see typechecker
		// errors like "StackGuardMultiplier not declared by package
		// sys."
		fs = newWithFileOverlaid(fs, "/src/runtime/internal/sys/zversion.go", []byte(fmt.Sprintf(`package sys;const DefaultGoroot = %q;const TheVersion = %q;const Goexperiment="";const StackGuardMultiplier=1`, goroot, runtime.Version())))
	}

	var oldPath string
	if _, isStdlib := stdlibPackagePaths[d.importPath]; isStdlib {
		oldPath = goroot // stdlib
	} else {
		oldPath = path.Join(gopath, "src", d.projectRoot) // non-stdlib
	}

	h.handlerShared.mu.Lock()
	h.fs.Bind(oldPath, fs, "/", vfs.BindAfter)
	h.handlerShared.mu.Unlock()

	return nil
}

func doDeps(pkg *build.Package, mode build.ImportMode, importPackage func(path, srcDir string, mode build.ImportMode) (*build.Package, error)) error {
	// Separate mutexes for each package import path.
	var musMu sync.Mutex
	mus := map[string]*sync.Mutex{}
	mu := func(path string) *sync.Mutex {
		musMu.Lock()
		mu, ok := mus[path]
		if !ok {
			mu = new(sync.Mutex)
			mus[path] = mu
		}
		musMu.Unlock()
		return mu
	}

	gate := make(chan struct{}, runtime.GOMAXPROCS(0)) // I/O concurrency limit
	type importKey struct {
		path, srcDir string
		mode         build.ImportMode
	}
	type importResult struct {
		pkg *build.Package
		err error
	}
	var importCacheMu sync.Mutex
	importCache := map[importKey]importResult{}
	cachedImportPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		importCacheMu.Lock()
		res, ok := importCache[importKey{path, srcDir, mode}]
		importCacheMu.Unlock()
		if ok {
			return res.pkg, res.err
		}

		gate <- struct{}{} // limit I/O concurrency
		defer func() { <-gate }()

		mu := mu(path)
		mu.Lock() // only try to import a path once
		defer mu.Unlock()

		importCacheMu.Lock()
		res, ok = importCache[importKey{path, srcDir, mode}]
		importCacheMu.Unlock()
		if !ok {
			res.pkg, res.err = importPackage(path, srcDir, mode)
			importCacheMu.Lock()
			importCache[importKey{path, srcDir, mode}] = res
			importCacheMu.Unlock()
		}
		return res.pkg, res.err
	}

	var seenMu sync.Mutex
	seen := map[string]struct{}{}

	var errs errorList
	var wg sync.WaitGroup
	var do func(pkg *build.Package)
	do = func(pkg *build.Package) {
		for _, path := range allPackageImportsSorted(pkg) {
			seenMu.Lock()
			if _, seen := seen[path]; seen {
				seenMu.Unlock()
				continue
			}
			seen[path] = struct{}{}
			seenMu.Unlock()

			if path == "C" {
				continue
			}
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				pkg, err := cachedImportPackage(path, pkg.Dir, mode)
				if err != nil {
					errs.add(err)
				}
				if pkg != nil {
					do(pkg)
				}
			}(path)
		}
	}
	do(pkg)
	wg.Wait()
	return errs.error()
}

func allPackageImportsSorted(pkg *build.Package) []string {
	uniq := map[string]struct{}{}
	for _, p := range pkg.Imports {
		uniq[p] = struct{}{}
	}
	for _, p := range pkg.TestImports {
		uniq[p] = struct{}{}
	}
	for _, p := range pkg.XTestImports {
		uniq[p] = struct{}{}
	}
	imps := make([]string, 0, len(uniq))
	for p := range uniq {
		imps = append(imps, p)
	}
	sort.Strings(imps)
	return imps
}

func isMultiplePackageError(err error) bool {
	_, ok := err.(*build.MultiplePackageError)
	return ok
}
