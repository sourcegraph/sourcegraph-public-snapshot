package buildserver

import (
	"context"
	"fmt"
	"go/build"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/golang/go-langserver/langserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// fetchTransitiveDepsOfFile fetches the transitive dependencies of
// the named Go file. A Go file's dependencies are the imports of its
// own package, plus all of its imports' imports, and so on.
//
// It adds fetched dependencies to its own file system overlay, and
// the returned depFiles should be passed onto the language server to
// add to its overlay.
func (h *BuildHandler) fetchTransitiveDepsOfFile(ctx context.Context, fileURI string) (err error) {
	parentSpan := opentracing.SpanFromContext(ctx)
	span := parentSpan.Tracer().StartSpan("xlang-go: fetch transitive dependencies",
		opentracing.Tags{"fileURI": fileURI},
		opentracing.ChildOf(parentSpan.Context()),
	)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	bctx := h.OverlayBuildContext(ctx, &build.Context{
		GOOS:     goos,
		GOARCH:   goarch,
		GOPATH:   gopath,
		GOROOT:   goroot,
		Compiler: gocompiler,
	}, false)

	bpkg, err := langserver.ContainingPackage(bctx, h.FilePath(fileURI))
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

		// If this package resolves to the same repo, then use any
		// imported package, even if it has errors. The errors would
		// be caused by the repo itself, not our dep fetching.
		//
		// TODO(sqs): if a package example.com/a imports
		// example.com/a/b and example.com/a/b lives in a separate
		// repo, then this will break. This is the case for some
		// azul3d packages, but it's rare.
		if langserver.PathHasPrefix(path, h.rootImportPath) {
			if pkg != nil {
				return pkg, nil
			}
			return nil, fmt.Errorf("package %q is inside of workspace root but failed to import: %s", path, err)
		}

		// Otherwise, it's an external dependency. Fetch the package
		// and try again.
		d, err := resolveImportPath(http.DefaultClient, path)
		if err != nil {
			return nil, err
		}

		// If this package resolves to the same repo, then don't fetch
		// it; it is already on disk. If we fetch it, we might end up
		// with multiple conflicting versions of the workspace's repo
		// overlaid on each other.
		if langserver.PathHasPrefix(d.projectRoot, h.rootImportPath) {
			return nil, fmt.Errorf("package %q is inside of workspace root, refusing to fetch remotely", path)
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
	if d.vcs != "git" {
		return fmt.Errorf("dependency at import path %q has unsupported VCS %q (clone URL is %q)", d.importPath, d.vcs, d.cloneURL)
	}

	rev := d.rev
	if rev == "" {
		rev = "HEAD"
	}

	cloneURL, err := url.Parse(d.cloneURL)
	if err != nil {
		return err
	}
	fs, err := NewDepRepoVFS(cloneURL, rev)
	if err != nil {
		return err
	}

	if _, isStdlib := stdlibPackagePaths[d.importPath]; isStdlib {
		fs = addSysZversionFile(fs)
	}

	var oldPath string
	if _, isStdlib := stdlibPackagePaths[d.importPath]; isStdlib {
		oldPath = goroot // stdlib
	} else {
		oldPath = path.Join(gopath, "src", d.projectRoot) // non-stdlib
	}

	h.HandlerShared.Mu.Lock()
	h.FS.Bind(oldPath, fs, "/", ctxvfs.BindAfter)
	h.HandlerShared.Mu.Unlock()

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

// NewDepRepoVFS returns a virtual file system interface for accessing
// the files in the specified (public) repo at the given commit.
//
// TODO(sqs): design a way for the Go build/lang server to access
// private repos. Private repos are currently only supported for the
// main workspace repo, not as dependencies.
var NewDepRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
	// Fast-path for GitHub repos, which we can fetch on-demand from
	// GitHub's repo .zip archive download endpoint.
	if cloneURL.Host == "github.com" {
		fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
		return vfsutil.NewGitHubRepoVFS(fullName, rev, "", true)
	}

	// Fall back to a full git clone for non-github.com repos.
	return &vfsutil.GitRepoVFS{
		CloneURL: cloneURL.String(),
		Rev:      rev,
	}, nil
}
