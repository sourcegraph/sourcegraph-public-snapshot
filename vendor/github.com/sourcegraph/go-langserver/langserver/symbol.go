package langserver

import (
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/buildutil"

	"github.com/neelance/parallel"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Query is a structured representation that is parsed from the user's
// raw query string.
type Query struct {
	Kind      lsp.SymbolKind
	Filter    FilterType
	File, Dir string
	Tokens    []string
}

// String converts the query back into a logically equivalent, but not strictly
// byte-wise equal, query string. It is useful for converting a modified query
// structure back into a query string.
func (q Query) String() string {
	s := ""
	switch q.Filter {
	case FilterExported:
		s = queryJoin(s, "is:exported")
	case FilterDir:
		s = queryJoin(s, fmt.Sprintf("%s:%s", q.Filter, q.Dir))
	default:
		// no filter.
	}
	if q.Kind != 0 {
		for kwd, kind := range keywords {
			if kind == q.Kind {
				s = queryJoin(s, kwd)
			}
		}
	}
	for _, token := range q.Tokens {
		s = queryJoin(s, token)
	}
	return s
}

// queryJoin joins the strings into "<s><space><e>" ensuring there is no
// trailing or leading whitespace at the end of the string.
func queryJoin(s, e string) string {
	return strings.TrimSpace(s + " " + e)
}

// ParseQuery parses a user's raw query string and returns a
// structured representation of the query.
func ParseQuery(q string) (qu Query) {
	// All queries are case insensitive.
	q = strings.ToLower(q)

	// Split the query into space-delimited fields.
	for _, field := range strings.Fields(q) {
		// Check if the field is a filter like `is:exported`.
		if strings.HasPrefix(field, "dir:") {
			qu.Filter = FilterDir
			qu.Dir = strings.TrimPrefix(field, "dir:")
			continue
		}
		if field == "is:exported" {
			qu.Filter = FilterExported
			continue
		}

		// Each field is split into tokens, delimited by periods or slashes.
		tokens := strings.FieldsFunc(field, func(c rune) bool {
			return c == '.' || c == '/'
		})
		for _, tok := range tokens {
			if kind, isKeyword := keywords[tok]; isKeyword {
				qu.Kind = kind
				continue
			}
			qu.Tokens = append(qu.Tokens, tok)
		}
	}
	return qu
}

type FilterType string

const (
	FilterExported FilterType = "exported"
	FilterDir      FilterType = "dir"
)

// keywords are keyword tokens that will be interpreted as symbol kind
// filters in the search query.
var keywords = map[string]lsp.SymbolKind{
	"package": lsp.SKPackage,
	"type":    lsp.SKClass,
	"method":  lsp.SKMethod,
	"field":   lsp.SKField,
	"func":    lsp.SKFunction,
	"var":     lsp.SKVariable,
	"const":   lsp.SKConstant,
}

// resultSorter is a utility struct for collecting, filtering, and
// sorting symbol results.
type resultSorter struct {
	Query
	results   []scoredSymbol
	resultsMu sync.Mutex
}

// scoredSymbol is a symbol with an attached search relevancy score.
// It is used internally by resultSorter.
type scoredSymbol struct {
	score int
	lsp.SymbolInformation
}

/*
 * sort.Interface methods
 */
func (s *resultSorter) Len() int { return len(s.results) }
func (s *resultSorter) Less(i, j int) bool {
	iscore, jscore := s.results[i].score, s.results[j].score
	if iscore == jscore {
		if s.results[i].ContainerName == s.results[j].ContainerName {
			if s.results[i].Name == s.results[j].Name {
				return s.results[i].Location.URI < s.results[j].Location.URI
			}
			return s.results[i].Name < s.results[j].Name
		}
		return s.results[i].ContainerName < s.results[j].ContainerName
	}
	return iscore > jscore
}
func (s *resultSorter) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}

// Collect is a thread-safe method that will record the passed-in
// symbol in the list of results if its score > 0.
func (s *resultSorter) Collect(si lsp.SymbolInformation) {
	s.resultsMu.Lock()
	score := score(s.Query, si)
	if score > 0 {
		sc := scoredSymbol{score, si}
		s.results = append(s.results, sc)
	}
	s.resultsMu.Unlock()
}

// Results returns the ranked list of SymbolInformation values.
func (s *resultSorter) Results() []lsp.SymbolInformation {
	res := make([]lsp.SymbolInformation, len(s.results))
	for i, s := range s.results {
		res[i] = s.SymbolInformation
	}
	return res
}

// score returns 0 for results that aren't matches. Results that are matches are assigned
// a positive score, which should be used for ranking purposes.
func score(q Query, s lsp.SymbolInformation) (scor int) {
	if q.Kind != 0 {
		if q.Kind != s.Kind {
			return 0
		}
	}
	name, container := strings.ToLower(s.Name), strings.ToLower(s.ContainerName)
	filename := strings.TrimPrefix(strings.ToLower(s.Location.URI), "file://")
	isVendor := strings.HasPrefix(filename, "vendor/") || strings.Contains(filename, "/vendor/")
	if q.Filter == FilterExported && isVendor {
		// is:exported excludes vendor symbols always.
		return 0
	}
	if q.File != "" && filename != q.File {
		// We're restricting results to a single file, and this isn't it.
		return 0
	}
	if len(q.Tokens) == 0 { // early return if empty query
		if isVendor {
			return 1 // lower score for vendor symbols
		} else {
			return 2
		}
	}
	for i, tok := range q.Tokens {
		tok := strings.ToLower(tok)
		if strings.HasPrefix(container, tok) {
			scor += 2
		}
		if strings.HasPrefix(name, tok) {
			scor += 3
		}
		if strings.Contains(filename, tok) && len(tok) >= 3 {
			scor++
		}
		if strings.HasPrefix(filepath.Base(filename), tok) && len(tok) >= 3 {
			scor += 2
		}
		if tok == name {
			if i == len(q.Tokens)-1 {
				scor += 50
			} else {
				scor += 5
			}
		}
		if tok == container {
			scor += 3
		}
	}
	if scor > 0 && !(strings.HasPrefix(filename, "vendor/") || strings.Contains(filename, "/vendor/")) {
		// boost for non-vendor symbols
		scor += 5
	}
	if scor > 0 && ast.IsExported(s.Name) {
		// boost for exported symbols
		scor++
	}
	return scor
}

// toSym returns a SymbolInformation value derived from values we get
// from the Go parser and doc packages.
func toSym(name, container string, kind lsp.SymbolKind, fs *token.FileSet, pos token.Pos) lsp.SymbolInformation {
	container = filepath.Base(container)
	if f := strings.Fields(container); len(f) > 0 {
		container = f[len(f)-1]
	}
	return lsp.SymbolInformation{
		Name:          name,
		Kind:          kind,
		Location:      goRangeToLSPLocation(fs, pos, pos+token.Pos(len(name))-1),
		ContainerName: container,
	}
}

// handleTextDocumentSymbol handles `textDocument/documentSymbol` requests for
// the Go language server.
func (h *LangHandler) handleTextDocumentSymbol(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.DocumentSymbolParams) ([]lsp.SymbolInformation, error) {
	f := strings.TrimPrefix(params.TextDocument.URI, "file://")
	return h.handleSymbol(ctx, conn, req, Query{File: f}, 0)
}

// handleSymbol handles `workspace/symbol` requests for the Go
// language server.
func (h *LangHandler) handleWorkspaceSymbol(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	q := ParseQuery(params.Query)
	if q.Filter == FilterDir {
		q.Dir = path.Join(h.init.RootImportPath, q.Dir)
	}
	return h.handleSymbol(ctx, conn, req, q, params.Limit)
}

func (h *LangHandler) handleSymbol(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, query Query, limit int) ([]lsp.SymbolInformation, error) {
	results := resultSorter{Query: query, results: make([]scoredSymbol, 0)}
	{
		fs := token.NewFileSet()
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

		par := parallel.NewRun(8)
		pkgs := buildutil.ExpandPatterns(bctx, []string{pkgPat})
		for pkg := range pkgs {
			// If we're restricting results to a single file or dir, ensure the
			// package dir matches to avoid doing unnecessary work.
			if results.Query.File != "" {
				filePkgPath := path.Dir(results.Query.File)
				if PathHasPrefix(filePkgPath, h.init.BuildContext.GOROOT) {
					filePkgPath = PathTrimPrefix(filePkgPath, h.init.BuildContext.GOROOT)
				} else {
					filePkgPath = PathTrimPrefix(filePkgPath, h.init.BuildContext.GOPATH)
				}
				filePkgPath = PathTrimPrefix(filePkgPath, "src")
				if !pathEqual(pkg, filePkgPath) {
					continue
				}
			}
			if results.Query.Filter == FilterDir && !pathEqual(pkg, results.Query.Dir) {
				continue
			}

			par.Acquire()
			go func(pkg string) {
				defer par.Release()
				// Prevent any uncaught panics from taking the
				// entire server down. For an example see
				// https://github.com/golang/go/issues/17788
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
				h.collectFromPkg(bctx, fs, pkg, rootPath, &results)
			}(pkg)
		}
		_ = par.Wait()
	}
	sort.Sort(&results)
	if len(results.results) > limit && limit > 0 {
		results.results = results.results[:limit]
	}

	return results.Results(), nil
}

// getPkgSyms returns the cached symbols for package pkg, if they
// exist. Otherwise, it returns nil.
func (h *LangHandler) getPkgSyms(pkg string) []lsp.SymbolInformation {
	h.pkgSymCacheMu.Lock()
	l, ok := h.pkgSymCache[pkg]
	h.pkgSymCacheMu.Unlock()
	if ok {
		symbolCacheTotal.WithLabelValues("hit").Inc()
	} else {
		symbolCacheTotal.WithLabelValues("miss").Inc()
	}
	return l
}

// setPkgSyms updates the cached symbols for package pkg.
func (h *LangHandler) setPkgSyms(pkg string, syms []lsp.SymbolInformation) {
	h.pkgSymCacheMu.Lock()
	if h.pkgSymCache == nil {
		h.pkgSymCache = map[string][]lsp.SymbolInformation{}
	}
	h.pkgSymCache[pkg] = syms
	h.pkgSymCacheMu.Unlock()
}

// collectFromPkg collects all the symbols from the specified package
// into the results. It uses LangHandler's package symbol cache to
// speed up repeated calls.
func (h *LangHandler) collectFromPkg(bctx *build.Context, fs *token.FileSet, pkg string, rootPath string, results *resultSorter) {
	pkgSyms := h.getPkgSyms(pkg)
	if pkgSyms == nil {
		buildPkg, err := bctx.Import(pkg, rootPath, 0)
		if err != nil {
			maybeLogImportError(pkg, err)
			return
		}

		astPkgs, err := parseDir(fs, bctx, buildPkg.Dir, nil, 0)
		if err != nil {
			log.Printf("failed to parse directory %s: %s", buildPkg.Dir, err)
			return
		}
		astPkg := astPkgs[buildPkg.Name]
		if astPkg == nil {
			if !strings.HasPrefix(buildPkg.ImportPath, "github.com/golang/go/misc/cgo/") {
				log.Printf("didn't find build package name %q in parsed AST packages %v", buildPkg.ImportPath, astPkgs)
			}
			return
		}
		// TODO(keegancsmith) Remove vendored doc/go once https://github.com/golang/go/issues/17788 is shipped
		docPkg := doc.New(astPkg, buildPkg.ImportPath, doc.AllDecls)

		// Emit decls
		for _, t := range docPkg.Types {
			if len(t.Decl.Specs) == 1 { // the type name is the first spec in type declarations
				pkgSyms = append(pkgSyms, toSym(t.Name, buildPkg.ImportPath, lsp.SKClass, fs, t.Decl.Specs[0].Pos()))
			} else { // in case there's some edge case where there's not 1 spec, fall back to the start of the declaration
				pkgSyms = append(pkgSyms, toSym(t.Name, buildPkg.ImportPath, lsp.SKClass, fs, t.Decl.TokPos))
			}

			for _, v := range t.Funcs {
				pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg.ImportPath, lsp.SKFunction, fs, v.Decl.Name.NamePos))
			}
			for _, v := range t.Methods {
				if results.Query.Filter == FilterExported && (!ast.IsExported(v.Name) || !ast.IsExported(t.Name)) {
					continue
				}
				pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg.ImportPath+" "+t.Name, lsp.SKMethod, fs, v.Decl.Name.NamePos))
			}
			for _, v := range t.Consts {
				for _, name := range v.Names {
					pkgSyms = append(pkgSyms, toSym(name, buildPkg.ImportPath, lsp.SKConstant, fs, v.Decl.TokPos))
				}
			}
			for _, v := range t.Vars {
				for _, name := range v.Names {
					pkgSyms = append(pkgSyms, toSym(name, buildPkg.ImportPath, lsp.SKField, fs, v.Decl.TokPos))
				}
			}
		}
		for _, v := range docPkg.Consts {
			for _, name := range v.Names {
				pkgSyms = append(pkgSyms, toSym(name, buildPkg.ImportPath, lsp.SKConstant, fs, v.Decl.TokPos))
			}
		}
		for _, v := range docPkg.Vars {
			for _, name := range v.Names {
				pkgSyms = append(pkgSyms, toSym(name, buildPkg.ImportPath, lsp.SKVariable, fs, v.Decl.TokPos))
			}
		}
		for _, v := range docPkg.Funcs {
			pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg.ImportPath, lsp.SKFunction, fs, v.Decl.Name.NamePos))
		}
		h.setPkgSyms(pkg, pkgSyms)
	}

	for _, sym := range pkgSyms {
		if results.Query.Filter == FilterExported && !ast.IsExported(sym.Name) {
			continue
		}
		results.Collect(sym)
	}
}

// parseDir mirrors parser.ParseDir, but uses the passed in build context's VFS. In other words,
// buildutil.parseFile : parser.ParseFile :: parseDir : parser.ParseDir
func parseDir(fset *token.FileSet, bctx *build.Context, path string, filter func(os.FileInfo) bool, mode parser.Mode) (pkgs map[string]*ast.Package, first error) {
	list, err := buildutil.ReadDir(bctx, path)
	if err != nil {
		return nil, err
	}

	pkgs = map[string]*ast.Package{}
	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".go") && (filter == nil || filter(d)) {
			filename := filepath.Join(path, d.Name())
			if src, err := buildutil.ParseFile(fset, bctx, nil, filepath.Join(path, d.Name()), filename, mode); err == nil {
				name := src.Name.Name
				pkg, found := pkgs[name]
				if !found {
					pkg = &ast.Package{
						Name:  name,
						Files: map[string]*ast.File{},
					}
					pkgs[name] = pkg
				}
				pkg.Files[filename] = src
			} else if first == nil {
				first = err
			}
		}
	}

	return
}

func maybeLogImportError(pkg string, err error) {
	_, isNoGoError := err.(*build.NoGoError)
	if !(isNoGoError || !isMultiplePackageError(err) || strings.HasPrefix(pkg, "github.com/golang/go/test/")) {
		log.Printf("skipping possible package %s: %s", pkg, err)
	}
}

var symbolCacheTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "golangserver",
	Subsystem: "symbol",
	Name:      "cache_request_total",
	Help:      "Count of requests to cache.",
}, []string{"type"})

func init() {
	prometheus.MustRegister(symbolCacheTotal)
}
