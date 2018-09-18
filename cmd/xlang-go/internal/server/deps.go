package server

import (
	"context"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/gosrc"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

type keyMutex struct {
	mu  sync.Mutex
	mus map[string]*sync.Mutex
}

// get returns a mutex unique to the given key.
func (k *keyMutex) get(key string) *sync.Mutex {
	k.mu.Lock()
	mu, ok := k.mus[key]
	if !ok {
		mu = &sync.Mutex{}
		k.mus[key] = mu
	}
	k.mu.Unlock()
	return mu
}

func newKeyMutex() *keyMutex {
	return &keyMutex{
		mus: map[string]*sync.Mutex{},
	}
}

type importKey struct {
	path, srcDir string
	mode         build.ImportMode
}

type importResult struct {
	pkg *build.Package
	err error
}

type depCache struct {
	importCacheMu sync.Mutex
	importCache   map[importKey]importResult

	// A mapping of package path -> direct import records
	collectReferences bool
	seenMu            sync.Mutex
	seen              map[string][]importRecord
	entryPackageDirs  []string
}

func newDepCache() *depCache {
	return &depCache{
		importCache: map[importKey]importResult{},
		seen:        map[string][]importRecord{},
	}
}

// fetchTransitiveDepsOfFile fetches the transitive dependencies of
// the named Go file. A Go file's dependencies are the imports of its
// own package, plus all of its imports' imports, and so on.
//
// It adds fetched dependencies to its own file system overlay, and
// the returned depFiles should be passed onto the language server to
// add to its overlay.
func (h *BuildHandler) fetchTransitiveDepsOfFile(ctx context.Context, fileURI lsp.DocumentURI, dc *depCache) (err error) {
	parentSpan := opentracing.SpanFromContext(ctx)
	span := parentSpan.Tracer().StartSpan("xlang-go: fetch transitive dependencies",
		opentracing.Tags{"fileURI": fileURI},
		opentracing.ChildOf(parentSpan.Context()),
	)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	bctx := h.lang.BuildContext(ctx)
	bpkg, err := langserver.ContainingPackage(bctx, h.FilePath(fileURI))
	if err != nil && !isMultiplePackageError(err) {
		return err
	}

	err = doDeps(bpkg, 0, dc, func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		return h.doFindPackage(ctx, bctx, path, srcDir, mode, dc)
	})
	return err
}

type findPkgKey struct {
	importPath string // e.g. "github.com/gorilla/mux"
	fromDir    string // e.g. "/gopath/src/github.com/kubernetes/kubernetes"
	mode       build.ImportMode
}

type findPkgValue struct {
	ready chan struct{} // closed to broadcast readiness
	bp    *build.Package
	err   error
}

func (h *BuildHandler) findPackageCached(ctx context.Context, bctx *build.Context, p, srcDir string, mode build.ImportMode) (*build.Package, error) {
	// bctx.FindPackage and loader.Conf does not have caching, and due to
	// vendor we need to repeat work. So what we do is normalise the
	// srcDir w.r.t. potential vendoring. This makes the assumption that
	// the underlying FS for bctx is always the same, which is currently a
	// correct assumption.
	//
	// Example: A project gh.com/p/r has a single vendor folder at
	// /goroot/gh.com/p/r/vendor. Both gh.com/p/r/foo and
	// gh.com/p/r/bar/baz use gh.com/gorilla/mux.
	// loader will then call both
	//
	//   findPackage(..., "gh.com/gorilla/mux", "/gopath/src/gh.com/p/r/foo", ...)
	//   findPackage(..., "gh.com/gorilla/mux", "/gopath/src/gh.com/p/r/bar/baz", ...)
	//
	// findPackage then starts from the directory and checks for any
	// potential vendor directories which contains
	// "gh.com/gorilla/mux". Given "/gopath/src/gh.com/p/r/foo" and
	// "/gopath/src/gh.com/p/r/bar/baz" may have different vendor dirs to
	// check, it can't cache this work.
	//
	// So instead of passing "/gopath/src/gh.com/p/r/bar/baz" we pass
	// "/gopath/src/gh.com/p/r" because we know the first vendor dir to
	// check is "/gopath/src/gh.com/p/r/vendor". This also means that
	// "/gopath/src/gh.com/p/r/bar/baz" and "/gopath/src/gh.com/p/r/foo"
	// get the same cache key findPkgKey{"gh.com/gorilla/mux", "/gopath/src/gh.com/p/r", 0}.
	if !build.IsLocalImport(p) && srcDir != "" {
		srcDirs := bctx.SrcDirs()
		isGoPathSrcDir := func(p string) bool {
			for _, d := range srcDirs {
				if p == d {
					return true
				}
			}
			return false
		}
		for !bctx.IsDir(path.Join(srcDir, "vendor", p)) && !isGoPathSrcDir(srcDir) && srcDir != goroot && srcDir != "/" {
			srcDir = path.Dir(srcDir)
		}
	}

	// We do single-flighting as well. conf.Loader does the same, but its
	// single-flighting is based on srcDir before it is normalised.
	k := findPkgKey{p, srcDir, mode}
	h.findPkgMu.Lock()
	if h.findPkg == nil {
		h.findPkg = make(map[findPkgKey]*findPkgValue)
	}
	v, ok := h.findPkg[k]
	if ok {
		h.findPkgMu.Unlock()
		<-v.ready

		return v.bp, v.err
	}

	v = &findPkgValue{ready: make(chan struct{})}
	h.findPkg[k] = v
	h.findPkgMu.Unlock()

	v.bp, v.err = h.findPackage(ctx, bctx, p, srcDir, mode)

	close(v.ready)
	return v.bp, v.err
}

// findPackage is a langserver.FindPackageFunc which integrates with the build
// server. It will fetch dependencies just in time.
func (h *BuildHandler) findPackage(ctx context.Context, bctx *build.Context, path, srcDir string, mode build.ImportMode) (*build.Package, error) {
	return h.doFindPackage(ctx, bctx, path, srcDir, mode, newDepCache())
}

// isUnderCanonicalImportPath tells if the given path is under the given root import path.
func isUnderRootImportPath(rootImportPath, path string) bool {
	return rootImportPath != "" && util.PathHasPrefix(path, rootImportPath)
}

func (h *BuildHandler) doFindPackage(ctx context.Context, bctx *build.Context, path, srcDir string, mode build.ImportMode, dc *depCache) (*build.Package, error) {
	bpkg, err := h.doFindPackage2(ctx, bctx, path, srcDir, mode, dc)

	// We couldn't find the package. If the package path is under the root when
	// both strings are lowercase, that is a good indicator the user may have
	// typo'd the case of their canonical import path. Their code would also
	// not compile in this case under Linux, but it would on a case-insensitive
	// filesystem like what Mac users typically have.
	if err != nil && isUnderRootImportPath(strings.ToLower(h.rootImportPath), strings.ToLower(path)) {
		err = fmt.Errorf("error importing %q: %s. This may be due to a case-sensitivity typo in your canonical import path comment. Found a similar root import path %q", path, err, h.rootImportPath)
	}
	if err != nil {
		// TODO(slimsag): Users do not have a way to see diagnostics, so if we
		// did not log this error here they would not be able to see it because
		// errors returned from this function go into diagnostics ultimately.
		log.Printf("error finding package %q: %s", path, err)
	}
	return bpkg, err
}

func (h *BuildHandler) doFindPackage2(ctx context.Context, bctx *build.Context, path, srcDir string, mode build.ImportMode, dc *depCache) (*build.Package, error) {
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
	if isUnderRootImportPath(h.rootImportPath, path) {
		if pkg != nil {
			return pkg, nil
		}
		return nil, fmt.Errorf("package %q is inside of workspace root but failed to import: %s", path, err)
	}

	// Otherwise, it's an external dependency. Fetch the package
	// and try again.
	d, err := gosrc.ResolveImportPath(h.cachingClient, path)
	if err != nil {
		return nil, err
	}

	// If this package resolves to the same repo, then don't fetch
	// it; it is already on disk. If we fetch it, we might end up
	// with multiple conflicting versions of the workspace's repo
	// overlaid on each other.
	if h.rootImportPath != "" && util.PathHasPrefix(d.ProjectRoot, h.rootImportPath) {
		return nil, fmt.Errorf("package %q is inside of workspace root, refusing to fetch remotely", path)
	}

	urlMu := h.depURLMutex.get(d.CloneURL)
	urlMu.Lock()
	defer urlMu.Unlock()

	// Check again after waiting.
	pkg, err = bctx.Import(path, srcDir, mode)
	if err == nil {
		return pkg, nil
	}

	// We may have a specific rev to use (from glide.lock)
	if rev := h.pinnedDep(ctx, d.ImportPath); rev != "" {
		d.Rev = rev
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
}

func (h *BuildHandler) fetchDep(ctx context.Context, d *gosrc.Directory) error {
	if d.VCS != "git" {
		return fmt.Errorf("dependency at import path %q has unsupported VCS %q (clone URL is %q)", d.ImportPath, d.VCS, d.CloneURL)
	}

	rev := d.Rev
	if rev == "" {
		rev = "HEAD"
	}
	cloneURL, err := url.Parse(d.CloneURL)
	if err != nil {
		return err
	}
	fs, err := NewDepRepoVFS(ctx, cloneURL, rev)
	if err != nil {
		return err
	}

	isStdlib := gosrc.IsStdlibPkg(d.ImportPath)
	if isStdlib {
		fs = addSysZversionFile(fs)
	}

	var oldPath string
	if isStdlib {
		oldPath = goroot // stdlib
	} else {
		oldPath = path.Join(gopath, "src", d.ProjectRoot) // non-stdlib
	}

	h.HandlerShared.Mu.Lock()
	h.FS.Bind(oldPath, fs, "/", ctxvfs.BindAfter)
	if !isStdlib {
		h.gopathDeps = append(h.gopathDeps, d)
	}
	h.HandlerShared.Mu.Unlock()

	return nil
}

func (h *BuildHandler) pinnedDep(ctx context.Context, pkg string) string {
	h.pinnedDepsOnce.Do(func() {
		h.HandlerShared.Mu.Lock()
		fs := h.FS
		root := h.RootFSPath
		h.HandlerShared.Mu.Unlock()

		// github.com/golang/dep is not widely used yet, but likely will in
		// the future. So we try it first.
		toml, err := ctxvfs.ReadFile(ctx, fs, path.Join(root, "Gopkg.lock"))
		if err == nil && len(toml) > 0 {
			h.pinnedDeps = loadGopkgLock(toml)
			return
		}

		// We assume glide.lock is in the top-level dir of the
		// repo. This assumption may not be valid in the future.
		yml, err := ctxvfs.ReadFile(ctx, fs, path.Join(root, "glide.lock"))
		if err == nil && len(yml) > 0 {
			h.pinnedDeps = loadGlideLock(yml)
			return
		}

		// Next we try load from Godeps. Note: We will mount the wrong
		// dependencies in these two strange cases:
		// 1. Different revisions for pkgs in the same repo.
		// 2. Using a pkg not in Godeps, but another pkg from the same repo is in Godeps
		// In both cases, we use the revision for the pkg we first try and fetch.
		b, err := ctxvfs.ReadFile(ctx, fs, path.Join(root, "Godeps/Godeps.json"))
		if err == nil && len(b) > 0 {
			h.pinnedDeps = loadGodeps(b)
			return
		}
	})
	return h.pinnedDeps.Find(pkg)
}

func doDeps(pkg *build.Package, mode build.ImportMode, dc *depCache, importPackage func(path, srcDir string, mode build.ImportMode) (*build.Package, error)) error {
	// Separate mutexes for each package import path.
	importPathMutex := newKeyMutex()

	gate := make(chan struct{}, runtime.GOMAXPROCS(0)) // I/O concurrency limit
	cachedImportPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		dc.importCacheMu.Lock()
		res, ok := dc.importCache[importKey{path, srcDir, mode}]
		dc.importCacheMu.Unlock()
		if ok {
			return res.pkg, res.err
		}

		gate <- struct{}{} // limit I/O concurrency
		defer func() { <-gate }()

		mu := importPathMutex.get(path)
		mu.Lock() // only try to import a path once
		defer mu.Unlock()

		dc.importCacheMu.Lock()
		res, ok = dc.importCache[importKey{path, srcDir, mode}]
		dc.importCacheMu.Unlock()
		if !ok {
			res.pkg, res.err = importPackage(path, srcDir, mode)
			dc.importCacheMu.Lock()
			dc.importCache[importKey{path, srcDir, mode}] = res
			dc.importCacheMu.Unlock()
		}
		return res.pkg, res.err
	}

	var errs errorList
	var wg sync.WaitGroup
	var do func(pkg *build.Package)
	do = func(pkg *build.Package) {
		dc.seenMu.Lock()
		if _, seen := dc.seen[pkg.Dir]; seen {
			dc.seenMu.Unlock()
			return
		}
		dc.seen[pkg.Dir] = []importRecord{}
		dc.seenMu.Unlock()

		for _, path := range allPackageImportsSorted(pkg) {
			if path == "C" {
				continue
			}
			wg.Add(1)
			parentPkg := pkg
			go func(path string) {
				defer wg.Done()
				pkg, err := cachedImportPackage(path, pkg.Dir, mode)
				if err != nil {
					errs.add(err)
				}
				if pkg != nil {
					if dc.collectReferences {
						dc.seenMu.Lock()
						dc.seen[parentPkg.Dir] = append(dc.seen[parentPkg.Dir], importRecord{pkg: parentPkg, imports: pkg})
						dc.seenMu.Unlock()
					}
					do(pkg)
				}
			}(path)
		}
	}
	do(pkg)
	if dc.collectReferences {
		dc.seenMu.Lock()
		dc.entryPackageDirs = append(dc.entryPackageDirs, pkg.Dir)
		dc.seenMu.Unlock()
	}
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

// FetchCommonDeps will fetch our common used dependencies. This is to avoid
// impacting the first ever typecheck we do in a repo since it will have to
// fetch the dependency from the internet.
func FetchCommonDeps() {
	// github.com/golang/go
	d, _ := gosrc.ResolveImportPath(http.DefaultClient, "time")
	u, _ := url.Parse(d.CloneURL)
	_, _ = NewDepRepoVFS(context.Background(), u, d.Rev)
}

// NewDepRepoVFS returns a virtual file system interface for accessing
// the files in the specified (public) repo at the given commit.
var NewDepRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
	// First check if we can clone from gitserver. gitserver automatically
	// clones missing repositories, so to prevent cloning unmanaged
	// repositories we first check to see if it is present.
	name := api.RepoURI(cloneURL.Host + cloneURL.Path)
	if cloned, _ := gitserver.DefaultClient.IsRepoCloned(ctx, name); cloned {
		repo := gitserver.Repo{Name: name}
		if commit, err := git.ResolveRevision(ctx, repo, nil, rev, nil); err == nil {
			return vfsutil.NewGitServer(name, commit), nil
		}
	}

	// Fast-path for GitHub repos, which we can fetch on-demand from
	// GitHub's repo .zip archive download endpoint.
	if cloneURL.Host == "github.com" {
		fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
		return vfsutil.NewGitHubRepoVFS(fullName, rev)
	}

	// Fall back to a full git clone for non-github.com repos.
	return &vfsutil.GitRepoVFS{
		CloneURL: cloneURL.String(),
		Rev:      rev,
	}, nil
}
