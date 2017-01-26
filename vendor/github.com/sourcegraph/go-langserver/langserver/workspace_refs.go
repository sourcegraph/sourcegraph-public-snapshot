package langserver

import (
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/tools/go/loader"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/langserver/internal/refs"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// workspaceReferencesTimeout is the timeout used for workspace/xreferences
// calls.
const workspaceReferencesTimeout = 15 * time.Second

func (h *LangHandler) handleWorkspaceReferences(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lspext.WorkspaceReferencesParams) ([]referenceInformation, error) {
	// TODO: Add support for the cancelRequest LSP method instead of using
	// hard-coded timeouts like this here.
	//
	// See: https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#cancelRequest
	ctx, cancel := context.WithTimeout(ctx, workspaceReferencesTimeout)
	defer cancel()
	rootPath := h.FilePath(h.init.RootPath)
	bctx := h.BuildContext(ctx)

	// Perform typechecking.
	var (
		findPackage        = h.getFindPackageFunc()
		fset               = token.NewFileSet()
		pkgs               []string
		unvendoredPackages = map[string]struct{}{}
	)
	for _, pkg := range listPkgsUnderDir(bctx, rootPath) {
		bpkg, err := findPackage(ctx, bctx, pkg, rootPath, build.FindOnly)
		if err != nil && !isMultiplePackageError(err) {
			log.Printf("skipping possible package %s: %s", pkg, err)
			continue
		}

		// If a dirs hint is present, only look for references created in those
		// directories.
		dirs, ok := params.Hints["dirs"]
		if ok {
			found := false
			for _, dir := range dirs.([]interface{}) {
				if "file://"+bpkg.Dir == dir.(string) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		unvendoredPackages[bpkg.ImportPath] = struct{}{}
		pkgs = append(pkgs, pkg)
	}
	if len(pkgs) == 0 {
		// occurs when the directory hint is present and matches no directories
		// at all.
		return []referenceInformation{}, nil
	}

	// Collect dependency references in the AfterTypeCheck phase. This enables
	// us to begin looking at packages as they are typechecked, instead of
	// waiting for all packages to be typechecked (which is IO bound).
	var (
		results = refResultSorter{results: make([]referenceInformation, 0)}
		wg      sync.WaitGroup
	)
	afterTypeCheck := func(pkg *loader.PackageInfo, files []*ast.File) {
		_, interested := unvendoredPackages[pkg.Pkg.Path()]
		if !interested {
			return
		}

		// Do not block the type-checker.
		wg.Add(1)
		go func() {
			// Prevent any uncaught panics from taking the entire server down.
			defer func() {
				wg.Done()
				_ = panicf(recover(), "%v for pkg %v", req.Method, pkg)
			}()

			err := h.workspaceRefsFromPkg(ctx, bctx, conn, params, fset, pkg, rootPath, &results)
			if err != nil {
				log.Printf("workspaceRefsFromPkg: %v: %v", pkg, err)
			}
		}()
	}

	// workspaceRefsTypecheck is ran inside its own goroutine because it can
	// block for longer than our context deadline.
	var err error
	done := make(chan struct{})
	go func() {
		// Prevent any uncaught panics from taking the entire server down.
		defer func() {
			_ = panicf(recover(), "%v for pkg %v", req.Method, pkgs)
		}()

		_, err = h.workspaceRefsTypecheck(ctx, bctx, conn, fset, pkgs, afterTypeCheck)

		// Wait for all worker goroutines to complete.
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	sort.Sort(&results) // sort to provide consistent results
	return results.results, nil
}

func (h *LangHandler) workspaceRefsTypecheck(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, fset *token.FileSet, pkgs []string, afterTypeCheck func(info *loader.PackageInfo, files []*ast.File)) (prog *loader.Program, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "workspaceRefsTypecheck")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Configure the loader.
	findPackage := h.getFindPackageFunc()
	var typeErrs []error
	conf := loader.Config{
		Fset: fset,
		TypeChecker: types.Config{
			DisableUnusedImportCheck: true,
			FakeImportC:              true,
			Error: func(err error) {
				typeErrs = append(typeErrs, err)
			},
		},
		Build:       bctx,
		AllowErrors: true,
		ParserMode:  parser.AllErrors | parser.ParseComments, // prevent parser from bailing out
		FindPackage: func(bctx *build.Context, importPath, fromDir string, mode build.ImportMode) (*build.Package, error) {
			// When importing a package, ignore any
			// MultipleGoErrors. This occurs, e.g., when you have a
			// main.go with "// +build ignore" that imports the
			// non-main package in the same dir.
			bpkg, err := findPackage(ctx, bctx, importPath, fromDir, mode)
			if err != nil && !isMultiplePackageError(err) {
				return bpkg, err
			}
			return bpkg, nil
		},
		AfterTypeCheck: func(pkg *loader.PackageInfo, files []*ast.File) {
			if err := ctx.Err(); err != nil {
				return
			}
			afterTypeCheck(pkg, files)
		},
	}
	for _, path := range pkgs {
		conf.Import(path)
	}

	// Load and typecheck the packages.
	prog, err = conf.Load()
	if err != nil && prog == nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Publish typechecking error diagnostics.
	diags, err := errsToDiagnostics(typeErrs, prog)
	if err != nil {
		return nil, err
	}
	if len(diags) > 0 {
		go func() {
			if err := h.publishDiagnostics(ctx, conn, diags); err != nil {
				log.Printf("warning: failed to send diagnostics: %s.", err)
			}
		}()
	}
	return prog, nil
}

// workspaceRefsFromPkg collects all the references made to dependencies from
// the specified package and returns the results.
func (h *LangHandler) workspaceRefsFromPkg(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, params lspext.WorkspaceReferencesParams, fs *token.FileSet, pkg *loader.PackageInfo, rootPath string, results *refResultSorter) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "workspaceRefsFromPkg")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("pkg", pkg)

	// Compute workspace references.
	findPackage := h.getFindPackageFunc()
	cfg := &refs.Config{
		FileSet:  fs,
		Pkg:      pkg.Pkg,
		PkgFiles: pkg.Files,
		Info:     &pkg.Info,
	}
	refsErr := cfg.Refs(func(r *refs.Ref) {
		symDesc, err := defSymbolDescriptor(ctx, bctx, rootPath, r.Def, findPackage)
		if err != nil {
			// Log the error, and flag it as one in the trace -- but do not
			// halt execution (hopefully, it is limited to a small subset of
			// the data).
			ext.Error.Set(span, true)
			err := fmt.Errorf("workspaceRefsFromPkg: failed to import %v: %v", r.Def.ImportPath, err)
			log.Println(err)
			span.SetTag("error", err.Error())
			return
		}
		if !symDesc.Contains(params.Query) {
			return
		}

		results.resultsMu.Lock()
		results.results = append(results.results, referenceInformation{
			Reference: goRangeToLSPLocation(fs, r.Start, r.End),
			Symbol:    symDesc,
		})
		results.resultsMu.Unlock()
	})
	if refsErr != nil {
		// Trace the error, but do not consider it a true error. In many cases
		// it is a problem with the user's code, not our workspace reference
		// finding code.
		span.SetTag("err", fmt.Sprintf("workspaceRefsFromPkg: workspace refs failed: %v: %v", pkg, refsErr))
	}
	return nil
}

func defSymbolDescriptor(ctx context.Context, bctx *build.Context, rootPath string, def refs.Def, findPackage FindPackageFunc) (*symbolDescriptor, error) {
	defPkg, err := findPackage(ctx, bctx, def.ImportPath, rootPath, build.FindOnly)
	if err != nil {
		return nil, err
	}

	// NOTE: fields must be kept in sync with symbol.go:symbolEqual
	desc := &symbolDescriptor{
		Vendor:      IsVendorDir(defPkg.Dir),
		Package:     defPkg.ImportPath,
		PackageName: def.PackageName,
		Recv:        "",
		Name:        "",
		ID:          "",
	}

	fields := strings.Fields(def.Path)
	switch {
	case len(fields) == 0:
		// reference to just a package
		desc.ID = fmt.Sprintf("%s", desc.Package)
	case len(fields) >= 2:
		desc.Recv = fields[0]
		desc.Name = fields[1]
		desc.ID = fmt.Sprintf("%s/-/%s/%s", desc.Package, desc.Recv, desc.Name)
	case len(fields) >= 1:
		desc.Name = fields[0]
		desc.ID = fmt.Sprintf("%s/-/%s", desc.Package, desc.Name)
	default:
		panic("invalid def.Path response from internal/refs")
	}
	return desc, nil
}

// refResultSorter is a utility struct for collecting, filtering, and
// sorting workspace reference results.
type refResultSorter struct {
	results   []referenceInformation
	resultsMu sync.Mutex
}

func (s *refResultSorter) Len() int      { return len(s.results) }
func (s *refResultSorter) Swap(i, j int) { s.results[i], s.results[j] = s.results[j], s.results[i] }
func (s *refResultSorter) Less(i, j int) bool {
	return s.results[i].Reference.URI < s.results[j].Reference.URI
}
