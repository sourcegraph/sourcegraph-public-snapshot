package langserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/refactor/importgraph"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/langserver/internal/tools"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// documentReferencesTimeout is the timeout used for textDocument/references
// calls.
const documentReferencesTimeout = 15 * time.Second

var streamExperiment = len(os.Getenv("STREAM_EXPERIMENT")) > 0

func (h *LangHandler) handleTextDocumentReferences(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	// TODO: Add support for the cancelRequest LSP method instead of using
	// hard-coded timeouts like this here.
	//
	// See: https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#cancelRequest
	ctx, cancel := context.WithTimeout(ctx, documentReferencesTimeout)
	defer cancel()

	// Begin computing the reverse import graph immediately, as this
	// occurs in the background and is IO-bound.
	reverseImportGraphC := h.reverseImportGraph(ctx, conn)

	fset, node, _, _, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no information.
		if _, ok := err.(*invalidNodeError); ok {
			return []lsp.Location{}, nil
		}
		return nil, err
	}

	// NOTICE: Code adapted from golang.org/x/tools/cmd/guru
	// referrers.go.

	obj := pkg.ObjectOf(node)
	if obj == nil {
		return nil, errors.New("references object not found")
	}

	// TODO(sqs): golang.org/x/tools/cmd/guru/referrers.go has some
	// other handling of obj == nil cases: type-switches, package
	// decls, and unresolved identifiers that we should adapt as well.
	if obj == nil {
		return nil, errors.New("object not found")
	}

	if obj.Pkg() == nil {
		if _, builtin := obj.(*types.Builtin); builtin {
			// We don't support builtin references due to the massive number
			// of references, so ignore the missing package error.
			return []lsp.Location{}, nil
		}
		return nil, fmt.Errorf("no package found for object %s", obj)
	}
	defpkg := strings.TrimSuffix(obj.Pkg().Path(), "_test")
	_, pkgLevel := classify(obj)

	bctx := h.BuildContext(ctx)
	pkgInWorkspace := func(path string) bool {
		return PathHasPrefix(path, h.init.RootImportPath)
	}

	var (
		// locsC receives the final collected references via
		// refStreamAndCollect.
		locsC = make(chan []lsp.Location)

		// refs is a stream of raw references found findReferences.
		refs = make(chan *ast.Ident)

		// findRefErr is non-nil if findReferences fails.
		findRefErr error
	)

	// Start a goroutine to read from the refs chan. It will read all the
	// refs until the chan is closed. It is responsible to stream the
	// references back to the client, as well as build up the final slice
	// which we return as the response.
	go func() {
		locsC <- refStreamAndCollect(ctx, conn, req, fset, refs)
		close(locsC)
	}()

	// Don't include decl if it is outside of workspace.
	if params.Context.IncludeDeclaration && PathHasPrefix(defpkg, h.init.RootImportPath) {
		refs <- &ast.Ident{NamePos: obj.Pos(), Name: obj.Name()}
	}

	// seen keeps track of already findReferenced packages. This allows us
	// to avoid doing extra work when we receive a successive import
	// graph.
	seen := make(map[string]bool)
	for reverseImportGraph := range reverseImportGraphC {
		// Find the set of packages in this workspace that depend on
		// defpkg. Only function bodies in those packages need
		// type-checking.
		var users map[string]bool
		if pkgLevel {
			users = map[string]bool{}
			for pkg := range reverseImportGraph[defpkg] {
				users[pkg] = true
			}
			users[defpkg] = true
		} else {
			users = reverseImportGraph.Search(defpkg)
		}

		// Anything in seen we have already collected references on,
		// so we only need to collect (users - seen).
		unseen := make(map[string]bool)
		for pkg := range users {
			if !seen[pkg] {
				unseen[pkg] = true
				seen[pkg] = true // need to mark for next loop
			}
		}
		if len(unseen) == 0 { // nothing to do
			continue
		}

		lconf := loader.Config{
			Fset:  fset,
			Build: bctx,
		}

		// The importgraph doesn't treat external test packages
		// as separate nodes, so we must use ImportWithTests.
		for path := range unseen {
			lconf.ImportWithTests(path)
		}

		findRefErr = findReferences(ctx, lconf, pkgInWorkspace, obj, refs)
	}

	// Tell refStreamAndCollect that we are done finding references. It
	// will then send the all the collected references to locsC.
	close(refs)
	locs := <-locsC

	// If a timeout does occur, we should know how effective the partial data is
	if ctx.Err() != nil {
		refTimeoutResults.Observe(float64(len(locs)))
		log.Printf("info: timeout during references for %s, found %d refs", defpkg, len(locs))
	}

	// If we find references then we can ignore findRefErr. It should only
	// be non-nil due to timeouts or our last findReferences doesn't find
	// the def.
	if len(locs) == 0 && findRefErr != nil {
		return nil, findRefErr
	}

	sortBySharedDirWithURI(params.TextDocument.URI, locs)

	// Technically we may be able to stop computing references sooner and
	// save RAM/CPU, but currently that would have two drawbacks:
	// * We can't stop the typechecking anyways
	// * We may return results that are not as interesting since sortBySharedDirWithURI won't see everything.
	if params.Context.XLimit > 0 && params.Context.XLimit < len(locs) {
		locs = locs[:params.Context.XLimit]
	}

	return locs, nil
}

// reverseImportGraph returns the reversed import graph for the workspace
// under the RootPath. Computing the reverse import graph is IO intensive, as
// such we may send down more than one import graph. The later a graph is
// sent, the more accurate it is. The channel will be closed, and the last
// graph sent is accurate. The reader does not have to read all the values.
func (h *LangHandler) reverseImportGraph(ctx context.Context, conn jsonrpc2.JSONRPC2) <-chan importgraph.Graph {
	// Ensure our buffer is big enough to prevent deadlock
	c := make(chan importgraph.Graph, 2)

	go func() {
		// This should always be related to the go import path for
		// this repo. For sourcegraph.com this means we share the
		// import graph across commits. We want this behaviour since
		// we assume that they don't change drastically across
		// commits.
		cacheKey := "importgraph:" + h.init.RootPath

		h.mu.Lock()
		tryCache := h.importGraph == nil
		once := h.importGraphOnce
		h.mu.Unlock()
		if tryCache {
			g := make(importgraph.Graph)
			if hit := h.cacheGet(ctx, conn, cacheKey, g); hit {
				// \o/
				c <- g
			}
		}

		parentCtx := ctx
		once.Do(func() {
			// Note: We use a background context since this
			// operation should not be cancelled due to an
			// individual request.
			span := startSpanFollowsFromContext(parentCtx, "BuildReverseImportGraph")
			ctx := opentracing.ContextWithSpan(context.Background(), span)
			defer span.Finish()

			bctx := h.BuildContext(ctx)
			findPackageWithCtx := h.getFindPackageFunc()
			findPackage := func(bctx *build.Context, importPath, fromDir string, mode build.ImportMode) (*build.Package, error) {
				return findPackageWithCtx(ctx, bctx, importPath, fromDir, mode)
			}
			g := tools.BuildReverseImportGraph(bctx, findPackage, h.FilePath(h.init.RootPath))
			h.mu.Lock()
			h.importGraph = g
			h.mu.Unlock()

			// Update cache in background
			go h.cacheSet(ctx, conn, cacheKey, g)
		})
		h.mu.Lock()
		// TODO(keegancsmith) h.importGraph may have been reset after once
		importGraph := h.importGraph
		h.mu.Unlock()
		c <- importGraph

		close(c)
	}()

	return c
}

// refStreamAndCollect returns all refs read in from chan until it is
// closed. While it is reading, it will also occasionaly stream out updates of
// the refs received so far.
func refStreamAndCollect(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, fset *token.FileSet, refs <-chan *ast.Ident) []lsp.Location {
	if !streamExperiment {
		var locs []lsp.Location
		for n := range refs {
			locs = append(locs, goRangeToLSPLocation(fset, n.Pos(), n.End()))
		}
		return locs
	}

	id := lsp.ID{
		Num:      req.ID.Num,
		Str:      req.ID.Str,
		IsString: req.ID.IsString,
	}
	initial := json.RawMessage(`[{"op":"replace","path":"","value":[]}]`)
	_ = conn.Notify(ctx, "$/partialResult", &lspext.PartialResultParams{
		ID:    id,
		Patch: &initial,
	})

	var (
		locs []lsp.Location
		pos  int
	)
	send := func() {
		if pos >= len(locs) {
			return
		}
		patch := make([]referenceAddOp, 0, len(locs)-pos)
		for _, l := range locs[pos:] {
			patch = append(patch, referenceAddOp{
				OP:    "add",
				Path:  "/-",
				Value: l,
			})
		}
		pos = len(locs)
		_ = conn.Notify(ctx, "$/partialResult", &lspext.PartialResultParams{
			ID: id,
			// We use referencePatch so the build server can rewrite URIs
			Patch: referencePatch(patch),
		})
	}

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case n, ok := <-refs:
			if !ok {
				// send a final update
				send()
				return locs
			}
			locs = append(locs, goRangeToLSPLocation(fset, n.Pos(), n.End()))
		case <-tick.C:
			send()
		}
	}
}

// findReferences will find all references to obj. It will only return
// references from packages in lconf.ImportPkgs.
func findReferences(ctx context.Context, lconf loader.Config, pkgInWorkspace func(string) bool, obj types.Object, refs chan<- *ast.Ident) error {
	allowErrors(&lconf)

	defpkg := strings.TrimSuffix(obj.Pkg().Path(), "_test")
	objposn := lconf.Fset.Position(obj.Pos())

	// The remainder of this function is somewhat tricky because it
	// operates on the concurrent stream of packages observed by the
	// loader's AfterTypeCheck hook.

	var (
		wg                sync.WaitGroup
		mu                sync.Mutex
		qobj              types.Object
		afterTypeCheckErr error
	)

	collectPkg := pkgInWorkspace
	if _, ok := lconf.ImportPkgs[defpkg]; !ok {
		// We have to typecheck defpkg, so just avoid references being collected.
		collectPkg = func(path string) bool {
			path = strings.TrimSuffix(path, "_test")
			return pkgInWorkspace(path) && path != defpkg
		}
		lconf.ImportWithTests(defpkg)
	}

	// Only typecheck pkgs which we can collect refs in, or the pkg our
	// object is defined in.
	lconf.TypeCheckFuncBodies = func(path string) bool {
		if ctx.Err() != nil {
			return false
		}

		path = strings.TrimSuffix(path, "_test")
		_, imported := lconf.ImportPkgs[path]
		return imported && (pkgInWorkspace(path) || path == defpkg)
	}

	// For efficiency, we scan each package for references
	// just after it has been type-checked. The loader calls
	// AfterTypeCheck (concurrently), providing us with a stream of
	// packages.
	lconf.AfterTypeCheck = func(info *loader.PackageInfo, files []*ast.File) {
		// AfterTypeCheck may be called twice for the same package due
		// to augmentation.

		defer clearInfoFields(info) // save memory

		wg.Add(1)
		defer wg.Done()

		pkg := strings.TrimSuffix(info.Pkg.Path(), "_test")

		// Only inspect packages that depend on the declaring package
		// (and thus were type-checked).
		if !lconf.TypeCheckFuncBodies(pkg) {
			return
		}

		// Record the query object and its package when we see
		// it. We can't reuse obj from the initial typecheck
		// because each go/loader Load invocation creates new
		// objects, and we need to test for equality later when we
		// look up refs.
		mu.Lock()
		if qobj == nil && pkg == defpkg {
			// Find the object by its position (slightly ugly).
			qobj = findObject(lconf.Fset, &info.Info, objposn)
			if qobj == nil {
				// It really ought to be there; we found it once
				// already.
				afterTypeCheckErr = fmt.Errorf("object at %s not found in package %s", objposn, defpkg)
			}
		}
		queryObj := qobj
		mu.Unlock()

		// Look for references to the query object. Only collect
		// those that are in this workspace.
		if queryObj != nil && collectPkg(pkg) {
			for id, obj := range info.Uses {
				if sameObj(queryObj, obj) {
					refs <- id
				}
			}
		}
	}

	// We don't use workgroup on this goroutine, since we want to return
	// early on context cancellation.
	done := make(chan struct{})
	go func() {
		// Prevent any uncaught panics from taking the entire server down.
		defer func() {
			close(done)
			_ = panicf(recover(), "findReferences")
		}()

		lconf.Load() // ignore error
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	// This should only wait in the case of the context being done. In
	// that case we are waiting for the currently running AfterTypeCheck
	// functions to finish.
	wg.Wait()

	if qobj == nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if afterTypeCheckErr != nil {
			// Only triggered by 1 specific error above (where we assign
			// afterTypeCheckErr), not any general loader error.
			return afterTypeCheckErr
		}
		return errors.New("query object not found during reloading")
	}

	return nil
}

// classify classifies objects by how far
// we have to look to find references to them.
func classify(obj types.Object) (global, pkglevel bool) {
	if obj.Exported() {
		if obj.Parent() == nil {
			// selectable object (field or method)
			return true, false
		}
		if obj.Parent() == obj.Pkg().Scope() {
			// lexical object (package-level var/const/func/type)
			return true, true
		}
	}
	// object with unexported named or defined in local scope
	return false, false
}

// allowErrors causes type errors to be silently ignored.
// (Not suitable if SSA construction follows.)
//
// NOTICE: Adapted from golang.org/x/tools.
func allowErrors(lconf *loader.Config) {
	ctxt := *lconf.Build // copy
	ctxt.CgoEnabled = false
	lconf.Build = &ctxt
	lconf.AllowErrors = true
	// AllErrors makes the parser always return an AST instead of
	// bailing out after 10 errors and returning an empty ast.File.
	lconf.ParserMode = parser.AllErrors
	lconf.TypeChecker.Error = func(err error) {}
}

// findObject returns the object defined at the specified position.
func findObject(fset *token.FileSet, info *types.Info, objposn token.Position) types.Object {
	good := func(obj types.Object) bool {
		if obj == nil {
			return false
		}
		posn := fset.Position(obj.Pos())
		return posn.Filename == objposn.Filename && posn.Offset == objposn.Offset
	}
	for _, obj := range info.Defs {
		if good(obj) {
			return obj
		}
	}
	for _, obj := range info.Implicits {
		if good(obj) {
			return obj
		}
	}
	return nil
}

func usesOf(queryObj types.Object, info *loader.PackageInfo) []*ast.Ident {
	var refs []*ast.Ident
	for id, obj := range info.Uses {
		if sameObj(queryObj, obj) {
			refs = append(refs, id)
		}
	}
	return refs
}

// same reports whether x and y are identical, or both are PkgNames
// that import the same Package.
func sameObj(x, y types.Object) bool {
	if x == y {
		return true
	}
	if x, ok := x.(*types.PkgName); ok {
		if y, ok := y.(*types.PkgName); ok {
			return x.Imported() == y.Imported()
		}
	}
	return false
}

func sortBySharedDirWithURI(uri string, locs []lsp.Location) {
	l := locationList{
		L: locs,
		D: make([]int, len(locs)),
	}
	// l.D[i] = number of shared directories between uri and l.L[i].URI
	for i := range l.L {
		u := l.L[i].URI
		var d int
		for i := 0; i < len(uri) && i < len(u) && uri[i] == u[i]; i++ {
			if u[i] == '/' {
				d++
			}
		}
		if u == uri {
			// Boost matches in the same uri
			d++
		}
		l.D[i] = d
	}
	sort.Sort(l)
}

type locationList struct {
	L []lsp.Location
	D []int
}

func (l locationList) Less(a, b int) bool {
	if l.D[a] != l.D[b] {
		return l.D[a] > l.D[b]
	}
	if x, y := path.Dir(l.L[a].URI), path.Dir(l.L[b].URI); x != y {
		return x < y
	}
	if l.L[a].URI != l.L[b].URI {
		return l.L[a].URI < l.L[b].URI
	}
	if l.L[a].Range.Start.Line != l.L[b].Range.Start.Line {
		return l.L[a].Range.Start.Line < l.L[b].Range.Start.Line
	}
	return l.L[a].Range.Start.Character < l.L[b].Range.Start.Character
}

func (l locationList) Swap(a, b int) {
	l.L[a], l.L[b] = l.L[b], l.L[a]
	l.D[a], l.D[b] = l.D[b], l.D[a]
}
func (l locationList) Len() int {
	return len(l.L)
}

var refTimeoutResults = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: "golangserver",
	Subsystem: "references",
	Name:      "timeout_references",
	Help:      "The number of references that were returned after a timeout.",
	// 0.01 is to capture no results
	Buckets: []float64{0.01, 1, 2, 32, 128, 1024},
})

func init() {
	prometheus.MustRegister(refTimeoutResults)
}
