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

	"golang.org/x/tools/go/loader"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/langserver/internal/refs"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleWorkspaceReferences(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lspext.WorkspaceReferencesParams) ([]lspext.ReferenceInformation, error) {
	// TODO(slimsag): respect params.Files which will make performance in any
	// moderately sized repository more bearable (right now these are really bad).

	rootPath := h.FilePath(h.init.RootPath)
	bctx := h.OverlayBuildContext(ctx, h.defaultBuildContext(), !h.init.NoOSFileSystemAccess)

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
	for _, pkg := range listPkgsUnderDir(bctx, rootPath) {
		// Ignore any vendor package so we can avoid scanning it for dependency
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
		pkgs = append(pkgs, pkg)
	}
	if len(pkgs) == 0 {
		// occurs when the directory hint is present and matches no directories
		// at all.
		return []lspext.ReferenceInformation{}, nil
	}
	prog, err := h.workspaceRefsTypecheck(ctx, bctx, conn, fset, pkgs)
	if err != nil {
		return nil, err
	}

	// Collect dependency references.
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
			err := h.workspaceRefsFromPkg(ctx, bctx, conn, params, fset, pkg, rootPath, &results)
			if err != nil {
				log.Printf("workspaceRefsFromPkg: %v: %v", pkg, err)
			}
		}(pkg)
	}
	_ = par.Wait()

	sort.Sort(&results) // sort to provide consistent results
	return results.results, nil
}

func (h *LangHandler) workspaceRefsTypecheck(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, fset *token.FileSet, pkgs []string) (prog *loader.Program, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "workspaceRefsTypecheck")
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

// workspaceRefsFromPkg collects all the references made to dependencies from
// the specified package and returns the results.
func (h *LangHandler) workspaceRefsFromPkg(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, params lspext.WorkspaceReferencesParams, fs *token.FileSet, pkg *loader.PackageInfo, rootPath string, results *refResultSorter) (err error) {
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
	cfg := &refs.Config{
		FileSet:  fs,
		Pkg:      pkg.Pkg,
		PkgFiles: pkg.Files,
		Info:     &pkg.Info,
	}
	refsErr := cfg.Refs(func(r *refs.Ref) {
		symDesc, err := defSymbolDescriptor(bctx, rootPath, r.Def)
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
		results.results = append(results.results, lspext.ReferenceInformation{
			Reference: goRangeToLSPLocation(fs, r.Pos, r.Pos), // TODO: internal/refs doesn't generate end positions
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

func defSymbolDescriptor(bctx *build.Context, rootPath string, def refs.Def) (lspext.SymbolDescriptor, error) {
	defPkg, err := bctx.Import(def.ImportPath, rootPath, build.FindOnly)
	if err != nil {
		return nil, err
	}

	desc := lspext.SymbolDescriptor{
		"vendor":      IsVendorDir(defPkg.Dir),
		"package":     defPkg.ImportPath,
		"packageName": def.PackageName,
		"recv":        "",
		"name":        "",
	}

	fields := strings.Fields(def.Path)
	switch {
	case len(fields) == 0:
		// reference to just a package
	case len(fields) >= 2:
		desc["recv"] = fields[0]
		desc["name"] = fields[1]
	case len(fields) >= 1:
		desc["name"] = fields[0]
	default:
		panic("invalid def.Path response from internal/refs")
	}
	return desc, nil
}

// refResultSorter is a utility struct for collecting, filtering, and
// sorting workspace reference results.
type refResultSorter struct {
	results   []lspext.ReferenceInformation
	resultsMu sync.Mutex
}

func (s *refResultSorter) Len() int      { return len(s.results) }
func (s *refResultSorter) Swap(i, j int) { s.results[i], s.results[j] = s.results[j], s.results[i] }
func (s *refResultSorter) Less(i, j int) bool {
	return s.results[i].Reference.URI < s.results[j].Reference.URI
}
