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
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// Query is a structured representation that is parsed from the user's
// raw query string.
type Query struct {
	Kind      lsp.SymbolKind
	Filter    FilterType
	File, Dir string
	Tokens    []string

	Symbol lspext.SymbolDescriptor
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

type symbolPair struct {
	lsp.SymbolInformation
	desc symbolDescriptor
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
	symbolPair
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
func (s *resultSorter) Collect(si symbolPair) {
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
func score(q Query, s symbolPair) (scor int) {
	if q.Kind != 0 {
		if q.Kind != s.Kind {
			return 0
		}
	}
	if q.Symbol != nil && !s.desc.Contains(q.Symbol) {
		return -1
	}
	name, container := strings.ToLower(s.Name), strings.ToLower(s.ContainerName)
	filename := strings.TrimPrefix(s.Location.URI, "file://")
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
func toSym(name string, bpkg *build.Package, recv string, kind lsp.SymbolKind, fs *token.FileSet, pos token.Pos) symbolPair {
	container := recv
	if container == "" {
		container = filepath.Base(bpkg.ImportPath)
	}

	var id string
	if recv == "" {
		id = fmt.Sprintf("%s/-/%s", path.Clean(bpkg.ImportPath), name)
	} else {
		id = fmt.Sprintf("%s/-/%s/%s", path.Clean(bpkg.ImportPath), recv, name)
	}

	return symbolPair{
		SymbolInformation: lsp.SymbolInformation{
			Name:          name,
			Kind:          kind,
			Location:      goRangeToLSPLocation(fs, pos, pos+token.Pos(len(name))-1),
			ContainerName: container,
		},
		// NOTE: fields must be kept in sync with workspace_refs.go:defSymbolDescriptor
		desc: symbolDescriptor{
			Vendor:      IsVendorDir(bpkg.Dir),
			Package:     path.Clean(bpkg.ImportPath),
			PackageName: bpkg.Name,
			Recv:        recv,
			Name:        name,
			ID:          id,
		},
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
func (h *LangHandler) handleWorkspaceSymbol(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	q := ParseQuery(params.Query)
	q.Symbol = params.Symbol
	if q.Filter == FilterDir {
		q.Dir = path.Join(h.init.RootImportPath, q.Dir)
	}
	return h.handleSymbol(ctx, conn, req, q, params.Limit)
}

func (h *LangHandler) handleSymbol(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, query Query, limit int) ([]lsp.SymbolInformation, error) {
	results := resultSorter{Query: query, results: make([]scoredSymbol, 0)}
	{
		rootPath := h.FilePath(h.init.RootPath)
		bctx := h.BuildContext(ctx)

		par := parallel.NewRun(8)
		for _, pkg := range listPkgsUnderDir(bctx, rootPath) {
			// If we're restricting results to a single file or dir, ensure the
			// package dir matches to avoid doing unnecessary work.
			if results.Query.File != "" {
				filePkgPath := path.Dir(results.Query.File)
				if PathHasPrefix(filePkgPath, bctx.GOROOT) {
					filePkgPath = PathTrimPrefix(filePkgPath, bctx.GOROOT)
				} else {
					filePkgPath = PathTrimPrefix(filePkgPath, bctx.GOPATH)
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
				// Prevent any uncaught panics from taking the
				// entire server down. For an example see
				// https://github.com/golang/go/issues/17788
				defer func() {
					par.Release()
					if r := recover(); r != nil {
						// Same as net/http
						const size = 64 << 10
						buf := make([]byte, size)
						buf = buf[:runtime.Stack(buf, false)]
						log.Printf("ignoring panic serving %v for pkg %v: %v\n%s", req.Method, pkg, r, buf)
						return
					}
				}()
				h.collectFromPkg(ctx, bctx, pkg, rootPath, &results)
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

type pkgSymResult struct {
	ready   chan struct{} // closed to broadcast readiness
	symbols []lsp.SymbolInformation
}

// collectFromPkg collects all the symbols from the specified package
// into the results. It uses LangHandler's package symbol cache to
// speed up repeated calls.
func (h *LangHandler) collectFromPkg(ctx context.Context, bctx *build.Context, pkg string, rootPath string, results *resultSorter) {
	symbols := h.symbolCache.Get(pkg, func() interface{} {
		findPackage := h.getFindPackageFunc()
		buildPkg, err := findPackage(ctx, bctx, pkg, rootPath, 0)
		if err != nil {
			maybeLogImportError(pkg, err)
			return nil
		}

		fs := token.NewFileSet()
		astPkgs, err := parseDir(fs, bctx, buildPkg.Dir, nil, 0)
		if err != nil {
			log.Printf("failed to parse directory %s: %s", buildPkg.Dir, err)
			return nil
		}
		astPkg := astPkgs[buildPkg.Name]
		if astPkg == nil {
			return nil
		}
		// TODO(keegancsmith) Remove vendored doc/go once https://github.com/golang/go/issues/17788 is shipped
		docPkg := doc.New(astPkg, buildPkg.ImportPath, doc.AllDecls)

		// Emit decls
		var pkgSyms []symbolPair
		for _, t := range docPkg.Types {
			if len(t.Decl.Specs) == 1 { // the type name is the first spec in type declarations
				pkgSyms = append(pkgSyms, toSym(t.Name, buildPkg, "", lsp.SKClass, fs, t.Decl.Specs[0].Pos()))
			} else { // in case there's some edge case where there's not 1 spec, fall back to the start of the declaration
				pkgSyms = append(pkgSyms, toSym(t.Name, buildPkg, "", lsp.SKClass, fs, t.Decl.TokPos))
			}

			for _, v := range t.Funcs {
				pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg, "", lsp.SKFunction, fs, v.Decl.Name.NamePos))
			}
			for _, v := range t.Methods {
				if results.Query.Filter == FilterExported && (!ast.IsExported(v.Name) || !ast.IsExported(t.Name)) {
					continue
				}
				pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg, t.Name, lsp.SKMethod, fs, v.Decl.Name.NamePos))
			}
			for _, v := range t.Consts {
				for _, name := range v.Names {
					pkgSyms = append(pkgSyms, toSym(name, buildPkg, "", lsp.SKConstant, fs, v.Decl.TokPos))
				}
			}
			for _, v := range t.Vars {
				for _, name := range v.Names {
					pkgSyms = append(pkgSyms, toSym(name, buildPkg, "", lsp.SKField, fs, v.Decl.TokPos))
				}
			}
		}
		for _, v := range docPkg.Consts {
			for _, name := range v.Names {
				pkgSyms = append(pkgSyms, toSym(name, buildPkg, "", lsp.SKConstant, fs, v.Decl.TokPos))
			}
		}
		for _, v := range docPkg.Vars {
			for _, name := range v.Names {
				pkgSyms = append(pkgSyms, toSym(name, buildPkg, "", lsp.SKVariable, fs, v.Decl.TokPos))
			}
		}
		for _, v := range docPkg.Funcs {
			pkgSyms = append(pkgSyms, toSym(v.Name, buildPkg, "", lsp.SKFunction, fs, v.Decl.Name.NamePos))
		}

		return pkgSyms
	})

	if symbols == nil {
		return
	}

	for _, sym := range symbols.([]symbolPair) {
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

// listPkgsUnderDir is buildutil.ExpandPattern(ctxt, []string{dir +
// "/..."}). The implementation is modified from the upstream
// buildutil.ExpandPattern so we can be much faster. buildutil.ExpandPattern
// looks at all directories under GOPATH if there is a `...` pattern. This
// instead only explores the directories under dir. In future
// buildutil.ExpandPattern may be more performant (there are TODOs for it).
func listPkgsUnderDir(ctxt *build.Context, dir string) []string {
	ch := make(chan string)

	var wg sync.WaitGroup
	for _, root := range ctxt.SrcDirs() {
		root := root
		wg.Add(1)
		go func() {
			allPackages(ctxt, root, dir, ch)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var pkgs []string
	for p := range ch {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	return pkgs
}

// We use a process-wide counting semaphore to limit
// the number of parallel calls to ReadDir.
var ioLimit = make(chan bool, 20)

// allPackages is from tools/go/buildutil. We don't use the exported method
// since it doesn't allow searching from a directory. We need from a specific
// directory for performance on large GOPATHs.
func allPackages(ctxt *build.Context, root, start string, ch chan<- string) {
	root = filepath.Clean(root) + string(os.PathSeparator)
	start = filepath.Clean(start) + string(os.PathSeparator)

	if strings.HasPrefix(root, start) {
		// If we are a child of start, we can just start at the
		// root. A concrete example of this happening is when
		// root=/goroot/src and start=/goroot
		start = root
	}

	if !strings.HasPrefix(start, root) {
		return
	}

	var wg sync.WaitGroup

	var walkDir func(dir string)
	walkDir = func(dir string) {
		// Avoid .foo, _foo, and testdata directory trees.
		base := filepath.Base(dir)
		if base == "" || base[0] == '.' || base[0] == '_' || base == "testdata" {
			return
		}

		pkg := filepath.ToSlash(strings.TrimPrefix(dir, root))

		// Prune search if we encounter any of these import paths.
		switch pkg {
		case "builtin":
			return
		}

		if pkg != "" {
			ch <- pkg
		}

		ioLimit <- true
		files, _ := buildutil.ReadDir(ctxt, dir)
		<-ioLimit
		for _, fi := range files {
			fi := fi
			if fi.IsDir() {
				wg.Add(1)
				go func() {
					walkDir(filepath.Join(dir, fi.Name()))
					wg.Done()
				}()
			}
		}
	}

	walkDir(start)
	wg.Wait()
}
