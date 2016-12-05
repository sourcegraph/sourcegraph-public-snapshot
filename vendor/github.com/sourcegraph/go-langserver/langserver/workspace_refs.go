package langserver

import (
	"context"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/langserver/internal/refs"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleWorkspaceReference(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lspext.WorkspaceReferenceParams) ([]lspext.ReferenceInformation, error) {
	rootPath := h.FilePath(h.init.RootPath)
	bctx := h.OverlayBuildContext(ctx, h.defaultBuildContext(), !h.init.NoOSFileSystemAccess)

	var pkgPat string
	if h.init.RootImportPath == "" {
		// Go stdlib (empty root import path)
		pkgPat = "..."
	} else {
		// All other Go packages.
		pkgPat = h.init.RootImportPath + "/..."
	}

	var parallelism int
	if envWorkspaceReferenceParallelism != "" {
		var err error
		parallelism, err = strconv.Atoi(envWorkspaceReferenceParallelism)
		if err != nil {
			return nil, err
		}
	} else {
		parallelism = runtime.NumCPU() / 4 // 1/4 CPU
	}
	if parallelism < 1 {
		parallelism = 1
	}

	// Perform typechecking.
	var (
		fset = token.NewFileSet()
		pkgs []string
	)
	for pkg := range buildutil.ExpandPatterns(bctx, []string{pkgPat}) {
		// Ignore any vendor package so we can avoid scanning it for external
		// references, per the workspace/reference spec. This saves us a
		// considerable amount of work.
		bpkg, err := bctx.Import(pkg, rootPath, build.FindOnly)
		if err != nil && !isMultiplePackageError(err) {
			log.Printf("skipping possible package %s: %s", pkg, err)
			continue
		}
		if IsVendorDir(bpkg.Dir) {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	prog, err := h.externalRefsTypecheck(ctx, bctx, conn, fset, pkgs)
	if err != nil {
		return nil, err
	}

	// Collect external references.
	results := refResultSorter{results: make([]lspext.ReferenceInformation, 0)}
	par := parallel.NewRun(parallelism)
	for _, pkg := range prog.Imported {
		par.Acquire()
		go func(pkg *loader.PackageInfo) {
			defer par.Release()
			// Prevent any uncaught panics from taking the entire server down.
			defer func() {
				if r := recover(); r != nil {
					// Same as net/http
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					log.Printf("ignoring panic serving %v for pkg %v: %v\n%s", req.Method, pkg, r, buf)
					return
				}
			}()
			err := h.externalRefsFromPkg(ctx, bctx, conn, fset, pkg, rootPath, &results)
			if err != nil {
				log.Printf("externalRefsFromPkg: %v: %v", pkg, err)
			}
		}(pkg)
	}
	_ = par.Wait()

	sort.Sort(&results) // sort to provide consistent results

	// TODO: We calculate all the results and then throw them away. If we ever
	// decide to begin using limiting, we can improve the performance of this
	// dramatically. For now, it lives in the spec just so that other
	// implementations are aware it may need to be done and to design with that
	// in mind.
	if len(results.results) > params.Limit && params.Limit > 0 {
		results.results = results.results[:params.Limit]
	}
	return results.results, nil
}

func (h *LangHandler) externalRefsTypecheck(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, fset *token.FileSet, pkgs []string) (prog *loader.Program, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "externalRefsTypecheck")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Configure the loader.
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
			bpkg, err := bctx.Import(importPath, fromDir, mode)
			if err != nil && !isMultiplePackageError(err) {
				return bpkg, err
			}
			return bpkg, nil
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

// externalRefsFromPkg collects all the external references from the specified
// package and returns the results.
func (h *LangHandler) externalRefsFromPkg(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, fs *token.FileSet, pkg *loader.PackageInfo, rootPath string, results *refResultSorter) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "externalRefsFromPkg")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("pkg", pkg)

	pkgInWorkspace := func(path string) bool {
		return PathHasPrefix(path, h.init.RootImportPath)
	}

	// Compute external references.
	cfg := &refs.Config{
		FileSet:  fs,
		Pkg:      pkg.Pkg,
		PkgFiles: pkg.Files,
		Info:     &pkg.Info,
	}
	refsErr := cfg.Refs(func(r *refs.Ref) {
		var defName, defContainerName string
		if fields := strings.Fields(r.Def.Path); len(fields) > 0 {
			defName = fields[0]
			defContainerName = strings.Join(fields[1:], " ")
		}
		if defContainerName == "" {
			defContainerName = r.Def.PackageName
		}

		defPkg, err := bctx.Import(r.Def.ImportPath, rootPath, build.FindOnly)
		if err != nil {
			// Log the error, and flag it as one in the trace -- but do not
			// halt execution (hopefully, it is limited to a small subset of
			// the data).
			ext.Error.Set(span, true)
			err := fmt.Errorf("externalRefsFromPkg: failed to import %v: %v", r.Def.ImportPath, err)
			log.Println(err)
			span.SetTag("error", err.Error())
			return
		}

		// If the symbol the reference is to is defined within this workspace,
		// exclude it. We only emit refs to symbols that are external to the
		// workspace.
		if pkgInWorkspace(defPkg.ImportPath) {
			return
		}

		results.resultsMu.Lock()
		results.results = append(results.results, lspext.ReferenceInformation{
			Name:          defName,
			ContainerName: defContainerName,
			URI:           "file://" + defPkg.Dir,
			Location:      goRangeToLSPLocation(fs, token.Pos(r.Position.Offset), token.Pos(r.Position.Offset+len(defName)-1)),
		})
		results.resultsMu.Unlock()
	})
	if refsErr != nil {
		// Trace the error, but do not consider it a true error. In many cases
		// it is a problem with the user's code, not our external reference
		// finding code.
		span.SetTag("err", fmt.Sprintf("externalRefsFromPkg: external refs failed: %v: %v", pkg, refsErr))
	}
	return nil
}

// refResultSorter is a utility struct for collecting, filtering, and
// sorting external reference results.
type refResultSorter struct {
	results   []lspext.ReferenceInformation
	resultsMu sync.Mutex
}

func (s *refResultSorter) Len() int      { return len(s.results) }
func (s *refResultSorter) Swap(i, j int) { s.results[i], s.results[j] = s.results[j], s.results[i] }
func (s *refResultSorter) Less(i, j int) bool {
	return s.results[i].Location.URI < s.results[j].Location.URI
}
