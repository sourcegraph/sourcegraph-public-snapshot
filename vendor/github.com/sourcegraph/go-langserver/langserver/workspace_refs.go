package langserver

import (
	"context"
	"fmt"
	"go/build"
	"go/token"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/buildutil"

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

	var paralellism int
	e := os.Getenv("WORKSPACE_REFERENCE_PARALLELISM")
	if e != "" {
		var err error
		paralellism, err = strconv.Atoi(e)
		if err != nil {
			return nil, err
		}
	} else {
		paralellism = runtime.NumCPU() / 4 // 1/4 CPU
	}
	if paralellism < 1 {
		paralellism = 1
	}

	results := refResultSorter{results: make([]lspext.ReferenceInformation, 0)}
	par := parallel.NewRun(paralellism)
	pkgs := buildutil.ExpandPatterns(bctx, []string{pkgPat})
	for pkg := range pkgs {
		par.Acquire()
		go func(pkg string) {
			defer par.Release()
			err := h.externalRefsFromPkg(ctx, bctx, conn, pkg, rootPath, &results)
			if err != nil {
				log.Printf("externalRefsFromPkg: %v: %v\n", pkg, err)
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

// externalRefsFromPkg collects all the external references from the specified
// package and returns the results.
func (h *LangHandler) externalRefsFromPkg(ctx context.Context, bctx *build.Context, conn JSONRPC2Conn, pkg string, rootPath string, results *refResultSorter) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "externalRefsFromPkg")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("pkg", pkg)

	// Import the package.
	buildPkg, err := bctx.Import(pkg, rootPath, 0)
	if err != nil {
		if !(strings.Contains(err.Error(), "no buildable Go source files") || strings.Contains(err.Error(), "found packages") || strings.HasPrefix(pkg, "github.com/golang/go/test/")) {
			return fmt.Errorf("skipping possible package %s: %s\n", pkg, err)
		}
		return nil
	}

	if strings.HasPrefix(buildPkg.Dir, "vendor/") || strings.Contains(buildPkg.Dir, "/vendor/") {
		// Per the workspace/reference docs:
		//
		// 	- Excluding any `URI` which is located within the workspace.
		// 	- Excluding any `Location` which is located in vendored code (e.g.
		// 	  `vendor/...` for Go, `node_modules/...` for JS, .tgz NPM packages, or
		// 	  .jar files for Java).
		//
		// This means that we do not need to consider vendor directories at all
		// since we do not emit references that are inside the same workspace
		// (vendor always is) and do not emit references to things that are
		// vendored. Thus, we can skip typechecking the entire vendor directory
		// which saves us a a lot of work.
		return nil
	}

	pkgInWorkspace := func(path string) bool {
		return PathHasPrefix(path, h.init.RootImportPath)
	}

	// Perform type checking.
	fs, prog, diags, err := h.cachedTypecheck(ctx, bctx, buildPkg)
	if err != nil {
		return err
	}
	if len(diags) > 0 {
		if err := h.publishDiagnostics(ctx, conn, diags); err != nil {
			return fmt.Errorf("sending diagnostics: %s", err)
		}
	}

	// Compute external references.
	cfg := &refs.Config{
		FileSet:  fs,
		Pkg:      prog.Package(pkg).Pkg,
		PkgFiles: prog.Package(pkg).Files,
		Info:     &prog.Package(pkg).Info,
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
		// Log the error, and flag it as one in the trace -- but do not
		// halt execution (hopefully, it is limited to a small subset of
		// the data, remember this is just external refs for one single
		// package).
		ext.Error.Set(span, true)
		err := fmt.Errorf("externalRefsFromPkg: external refs failed: %v: %v", pkg, refsErr)
		log.Println(err)
		span.SetTag("err", err.Error())
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
