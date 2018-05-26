package langserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
	"golang.org/x/tools/refactor/importgraph"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/go-langserver/pkg/tools"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleTextDocumentReferences(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	if !util.IsURI(params.TextDocument.URI) {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: fmt.Sprintf("textDocument/references not yet supported for out-of-workspace URI (%q)", params.TextDocument.URI),
		}
	}

	// Begin computing the reverse import graph immediately, as this
	// occurs in the background and is IO-bound.
	reverseImportGraphC := h.reverseImportGraph(ctx, conn)

	fset, node, _, _, pkg, _, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
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
		if h.init.RootImportPath == "" {
			return true
		}
		return util.PathHasPrefix(path, h.init.RootImportPath)
	}

	// findRefCtx is used in the findReferences function. It has its own
	// context so we can stop finding references once we have reached our
	// limit.
	findRefCtx, stop := context.WithCancel(ctx)
	defer stop()

	var (
		// locsC receives the final collected references via
		// refStreamAndCollect.
		locsC = make(chan []lsp.Location)

		// refs is a stream of raw references found by findReferences or findReferencesPkgLevel.
		refs = make(chan *ast.Ident)

		// findRefErr is non-nil if findReferences fails.
		findRefErr error
	)

	// Start a goroutine to read from the refs chan. It will read all the
	// refs until the chan is closed. It is responsible to stream the
	// references back to the client, as well as build up the final slice
	// which we return as the response.
	go func() {
		locsC <- refStreamAndCollect(ctx, conn, req, fset, refs, params.Context.XLimit, stop)
		close(locsC)
	}()

	// Don't include decl if it is outside of workspace.
	if params.Context.IncludeDeclaration && util.PathHasPrefix(defpkg, h.init.RootImportPath) {
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
			// We need to check all packages that import defpkg.
			users = map[string]bool{}
			for pkg := range reverseImportGraph[defpkg] {
				users[pkg] = true
			}
			// We also need to check defpkg itself, and its xtests.
			// For the reverse graph packages, we process xtests with the main package.
			// defpkg gets special handling; we must distinguish between in-package vs out-of-package.
			// To make the control flow in findReferencesPkgLevel simpler, add defpkg and defpkg xtest placeholders.
			// Use "!test" instead of "_test" because "!" is not a valid character in an import path.
			// (More precisely, it is not guaranteed to be a valid character in an import path,
			// so it is unlikely that it will be in use. See https://golang.org/ref/spec#Import_declarations.)
			users[defpkg] = true
			users[defpkg+"!test"] = true
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

		if pkgLevel {
			// pkgLevel queries can be done syntactically instead of semantically,
			// which is much faster. See https://golang.org/cl/97800/.
			findRefErr = h.findReferencesPkgLevel(findRefCtx, bctx, fset, unseen, pkgInWorkspace, obj, refs)
		} else {
			lconf := loader.Config{
				Fset:  fset,
				Build: bctx,
			}

			// The importgraph doesn't treat external test packages
			// as separate nodes, so we must use ImportWithTests.
			for path := range unseen {
				lconf.ImportWithTests(path)
			}

			findRefErr = findReferences(findRefCtx, lconf, pkgInWorkspace, obj, refs)
		}
		if findRefCtx.Err() != nil {
			// If we are canceled, cancel loop early
			break
		}
	}

	// Tell refStreamAndCollect that we are done finding references. It
	// will then send the all the collected references to locsC.
	close(refs)
	locs := <-locsC

	// If we find references then we can ignore findRefErr. It should only
	// be non-nil due to timeouts or our last findReferences doesn't find
	// the def.
	if len(locs) == 0 && findRefErr != nil {
		return nil, findRefErr
	}

	if locs == nil {
		locs = []lsp.Location{}
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
		cacheKey := "importgraph:" + string(h.init.Root())

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
			g := tools.BuildReverseImportGraph(bctx, findPackage, h.FilePath(h.init.Root()))
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
// closed. While it is reading, it will also occasionally stream out updates of
// the refs received so far.
func refStreamAndCollect(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, fset *token.FileSet, refs <-chan *ast.Ident, limit int, stop func()) []lsp.Location {
	if limit == 0 {
		// If we don't have a limit, just set it to a value we should never exceed
		limit = math.MaxInt32
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
			if len(locs) >= limit {
				stop()
				continue
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
	// Bail out early if the context is canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}

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
			_ = util.Panicf(recover(), "findReferences")
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

// findReferencesPkgLevel finds all references to obj.
// It only returns references from packages in users.
// It is the analogue of globalReferrersPkgLevel
// from golang.org/x/tools/cmd/guru/referrers.go.
func (h *LangHandler) findReferencesPkgLevel(ctx context.Context, bctx *build.Context, fset *token.FileSet, users map[string]bool, pkgInWorkspace func(string) bool, obj types.Object, refs chan<- *ast.Ident) error {
	// findReferencesPkgLevel uses go/ast and friends instead of go/types.
	// This affords a considerable performance benefit.
	// It comes at the cost of some code complexity.
	//
	// Here's a high level summary.
	//
	// The goal is to find references to the query object p.Q.
	// There are several possible scenarios, each handled differently.
	//
	// 1. We are looking in a package other than p, and p is not dot-imported.
	//    This is the simplest case. Q must be referred to as n.Q,
	//    where n is the name under which p is imported.
	//    We look at all imports of p to gather all names under which it is imported.
	//    (In the typical case, it is imported only once, under its default name.)
	//    Then we look at all selector expressions and report any matches.
	//
	// 2. We are looking in a package other than p, and p is dot-imported.
	//    In this case, Q will be referred to just as Q.
	//    Furthermore, go/ast's object resolution will not be able to resolve
	//    Q to any other object, unlike any local (file- or function- or block-scoped) object.
	//    So we look at all matching identifiers and report all unresolvable ones.
	//
	// 3. We are looking in package p.
	//    (Care must be taken to separate p and p_test (an xtest package),
	//    and make sure that they are treated as separate packages.)
	//    In this case, we give go/ast the entire package for object resolution,
	//    instead of going file by file.
	//    We then iterate over all identifiers that resolve to the query object.
	//    (The query object itself has already been reported, so we don't re-report it.)
	//
	// We always skip all files that don't contain the string Q, as they cannot be
	// relevant to finding references to Q.
	//
	// We parse all files leniently. In the presence of parsing errors, results are best-effort.

	defpkg := strings.TrimSuffix(obj.Pkg().Path(), "_test") // package x_test actually has package name x
	defpkg = imports.VendorlessPath(defpkg)

	defname := obj.Pkg().Name()                    // name of the defining package of the query object, used for resolving imports that use import path only (common case)
	isxtest := strings.HasSuffix(defname, "_test") // indicates whether the query object is defined in an xtest package

	name := obj.Name()
	namebytes := []byte(name)          // byte slice version of query object name, for early filtering
	objpos := fset.Position(obj.Pos()) // position of query object, used to prevent re-emitting original decl

	find := h.getFindPackageFunc()

	var reterr error
	sema := make(chan struct{}, 20) // counting semaphore to limit I/O concurrency
	var wg sync.WaitGroup
	for u := range users {
		// Bail out early if the context is canceled
		if err := ctx.Err(); err != nil {
			reterr = err
			// Don't "return err" here;
			// doing so would allow the caller to close the refs channel,
			// which might cause a panic if there are existing searches in flight.
			// Instead, just decline to start any new goroutines.
			break
		}

		u := u                                    // redeclare u to avoid range races
		uIsXTest := strings.HasSuffix(u, "!test") // indicates whether this package is the special defpkg xtest package
		u = strings.TrimSuffix(u, "!test")

		// pkgInWorkspace is cheap and usually false; check it before firing up a goroutine.
		if !pkgInWorkspace(u) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Resolve package.
			// TODO: is fromDir == "" correct?
			sema <- struct{}{} // acquire token
			pkg, err := find(ctx, bctx, u, "", build.IgnoreVendor)
			<-sema // release token
			if err != nil {
				return
			}

			// If we're not in the query package,
			// the object is in another package regardless,
			// so we want to process all files.
			// If we are in the query package,
			// we want to only process the files that are
			// part of that query package;
			// that set depends on whether the query package itself is an xtest.
			inQueryPkg := u == defpkg && isxtest == uIsXTest
			var files []string
			if !inQueryPkg || !isxtest {
				files = append(files, pkg.GoFiles...)
				files = append(files, pkg.TestGoFiles...)
				files = append(files, pkg.CgoFiles...) // use raw cgo files, as we're only parsing
			}
			if !inQueryPkg || isxtest {
				files = append(files, pkg.XTestGoFiles...)
			}

			if len(files) == 0 {
				return
			}

			var deffiles map[string]*ast.File // set of files that are part of this package, for inQueryPkg only
			if inQueryPkg {
				deffiles = make(map[string]*ast.File)
			}

			buf := new(bytes.Buffer) // reusable buffer for reading files

			for _, file := range files {
				if !buildutil.IsAbsPath(bctx, file) {
					file = buildutil.JoinPath(bctx, pkg.Dir, file)
				}
				buf.Reset()
				sema <- struct{}{} // acquire token
				src, err := readFile(bctx, file, buf)
				<-sema // release token
				if err != nil {
					continue
				}

				// Fast path: If the object's name isn't present anywhere in the source, ignore the file.
				if !bytes.Contains(src, namebytes) {
					continue
				}

				if inQueryPkg {
					// If we're in the query package, we defer final processing until we have
					// parsed all of the candidate files in the package.
					// Best effort; allow errors and use what we can from what remains.
					f, _ := parser.ParseFile(fset, file, src, parser.AllErrors)
					if f != nil {
						deffiles[file] = f
					}
					continue
				}

				// We aren't in the query package. Go file by file.

				// Parse out only the imports, to check whether the defining package
				// was imported, and if so, under what names.
				// Best effort; allow errors and use what we can from what remains.
				f, _ := parser.ParseFile(fset, file, src, parser.ImportsOnly|parser.AllErrors)
				if f == nil {
					continue
				}

				// pkgnames is the set of names by which defpkg is imported in this file.
				// (Multiple imports in the same file are legal but vanishingly rare.)
				pkgnames := make([]string, 0, 1)
				var isdotimport bool
				for _, imp := range f.Imports {
					path, err := strconv.Unquote(imp.Path.Value)
					if err != nil || path != defpkg {
						continue
					}
					switch {
					case imp.Name == nil:
						pkgnames = append(pkgnames, defname)
					case imp.Name.Name == ".":
						isdotimport = true
					default:
						pkgnames = append(pkgnames, imp.Name.Name)
					}
				}
				if len(pkgnames) == 0 && !isdotimport {
					// Defining package not imported, bail.
					continue
				}

				// Re-parse the entire file.
				// Parse errors are ok; we'll do the best we can with a partial AST, if we have one.
				f, _ = parser.ParseFile(fset, file, src, parser.AllErrors)
				if f == nil {
					continue
				}

				// Walk the AST looking for references.
				ast.Inspect(f, func(n ast.Node) bool {
					// Check selector expressions.
					// If the selector matches the target name,
					// and the expression is one of the names
					// that the defining package was imported under,
					// then we have a match.
					if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == name {
						if id, ok := sel.X.(*ast.Ident); ok {
							for _, n := range pkgnames {
								if n == id.Name {
									refs <- sel.Sel
									// Don't recurse further, to avoid duplicate entries
									// from the dot import check below.
									return false
								}
							}
						}
					}
					// Dot imports are special.
					// Objects imported from the defining package are placed in the package scope.
					// go/ast does not resolve them to an object.
					// At all other scopes (file, local), go/ast can do the resolution.
					// So we're looking for object-free idents with the right name.
					// The only other way to get something with the right name at the package scope
					// is to *be* the defining package. We handle that case separately (inQueryPkg).
					if isdotimport {
						if id, ok := n.(*ast.Ident); ok && id.Obj == nil && id.Name == name {
							refs <- id
							return false
						}
					}
					return true
				})
			}

			// If we're in the query package, we've now collected all the files in the package.
			// (Or at least the ones that might contain references to the object.)
			if inQueryPkg {
				// Bundle the files together into a package.
				// This does package-level object resolution.
				pkg, _ := ast.NewPackage(fset, deffiles, nil, nil)
				// Look up the query object; we know that it is defined in the package scope.
				pkgobj := pkg.Scope.Objects[name]
				if pkgobj == nil {
					panic("missing defpkg object for " + defpkg + "." + name)
				}
				// Find all references to the query object.
				ast.Inspect(pkg, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok {
						// Check both that this is a reference to the query object
						// and that it is not the query object itself;
						// the query object itself was already emitted.
						if id.Obj == pkgobj && objpos != fset.Position(id.Pos()) {
							refs <- id
							return false
						}
					}
					return true
				})
				deffiles = nil // allow GC
			}
		}()
	}

	wg.Wait()

	return reterr
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

// readFile is like ioutil.ReadFile, but
// it goes through the virtualized build.Context.
// If non-nil, buf must have been reset.
func readFile(ctxt *build.Context, filename string, buf *bytes.Buffer) ([]byte, error) {
	rc, err := buildutil.OpenFile(ctxt, filename)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	if buf == nil {
		buf = new(bytes.Buffer)
	}
	if _, err := io.Copy(buf, rc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
