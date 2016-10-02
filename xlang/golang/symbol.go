package golang

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
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/buildutil"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

type query struct {
	kind   lsp.SymbolKind
	tokens []string
}

var keywords = map[string]lsp.SymbolKind{
	"package": lsp.SKPackage,
	"type":    lsp.SKClass,
	"method":  lsp.SKMethod,
	"field":   lsp.SKField,
	"func":    lsp.SKFunction,
	"var":     lsp.SKVariable,
	"const":   lsp.SKConstant,
}

var tokenizer = regexp.MustCompile(`[\.\s\/\:]+`)

func parseQuery(q string) query {
	var qu query
	toks := tokenizer.Split(strings.ToLower(q), -1)
	for _, tok := range toks {
		if kind, isKeyword := keywords[tok]; isKeyword {
			qu.kind = kind
		} else {
			qu.tokens = append(qu.tokens, tok)
		}
	}
	return qu
}

// score returns 0 for results that aren't matches. Results that are matches are assigned
// a positive score, which should be used for ranking purposes.
func score(q query, s lsp.SymbolInformation) (scor int) {
	if q.kind != 0 {
		if q.kind != s.Kind {
			return 0
		}
	}
	name, container := strings.ToLower(s.Name), strings.ToLower(s.ContainerName)
	filename := strings.TrimPrefix(strings.ToLower(s.Location.URI), "file://")
	if len(q.tokens) == 0 {
		return 1
	}
	for i, tok := range q.tokens {
		tok := strings.ToLower(tok)
		if strings.HasPrefix(container, tok) {
			scor += 2
		}
		if strings.HasPrefix(name, tok) {
			scor += 3
		}
		if strings.Contains(filename, tok) && len(tok) >= 3 {
			scor += 1
		}
		if strings.HasPrefix(filepath.Base(filename), tok) && len(tok) >= 3 {
			scor += 2
		}
		if tok == name {
			if i == len(q.tokens)-1 {
				scor += 50
			} else {
				scor += 5
			}
		}
		if tok == container {
			scor += 3
		}
	}
	return scor
}

type scoredSymbol struct {
	score int
	lsp.SymbolInformation
}

type resultSorter struct {
	query
	results []scoredSymbol
}

func (s *resultSorter) Len() int { return len(s.results) }
func (s *resultSorter) Less(i, j int) bool {
	iscore, jscore := s.results[i].score, s.results[j].score
	if iscore == jscore {
		if s.results[i].ContainerName == s.results[j].ContainerName {
			return s.results[i].Name < s.results[j].Name
		}
		return s.results[i].ContainerName < s.results[j].ContainerName
	}
	return iscore > jscore
}
func (s *resultSorter) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}
func (s *resultSorter) Collect(si lsp.SymbolInformation) {
	score := score(s.query, si)
	if score > 0 {
		sc := scoredSymbol{score, si}
		s.results = append(s.results, sc)
	}
}
func (s *resultSorter) Results() []lsp.SymbolInformation {
	res := make([]lsp.SymbolInformation, len(s.results))
	for i, s := range s.results {
		res[i] = s.SymbolInformation
	}
	return res
}

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

func (h *LangHandler) handleSymbol(ctx context.Context, conn jsonrpc2Conn, req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	results := resultSorter{query: parseQuery(params.Query), results: make([]scoredSymbol, 0)}
	{
		fs := token.NewFileSet()
		rootPath := h.filePath(h.init.RootPath)
		bctx := h.overlayBuildContext(ctx, h.defaultBuildContext(), !h.init.NoOSFileSystemAccess)
		rootpkg, err := filepath.Rel(filepath.Join(bctx.GOPATH, "src"), rootPath)
		if err != nil {
			return nil, fmt.Errorf("workspace root path was not relative to $GOPATH/src: %s", err)
		}
		pkgs := buildutil.ExpandPatterns(bctx, []string{fmt.Sprintf("%s/...", rootpkg)})
		for pkg, _ := range pkgs {
			buildPkg, err := bctx.Import(pkg, rootPath, 0)
			if err != nil {
				log.Printf("skipping possible package %s: %s", pkg, err)
				continue
			}

			astPkgs, err := parseDir(fs, bctx, buildPkg.Dir, nil, 0)
			if err != nil {
				log.Printf("failed to parse directory %s: %s", buildPkg.Dir, err)
				continue
			}
			astPkg := astPkgs[buildPkg.Name]
			if astPkg == nil {
				log.Printf("didn't find build package name %q in parsed AST packages %v", buildPkg.ImportPath, astPkgs)
				continue
			}
			docPkg := doc.New(astPkg, buildPkg.ImportPath, doc.AllDecls)

			// Emit decls
			for _, t := range docPkg.Types {
				results.Collect(toSym(t.Name, buildPkg.ImportPath, lsp.SKClass, fs, t.Decl.TokPos))

				for _, v := range t.Funcs {
					results.Collect(toSym(v.Name, buildPkg.ImportPath, lsp.SKFunction, fs, v.Decl.Name.NamePos))
				}
				for _, v := range t.Methods {
					results.Collect(toSym(v.Name, buildPkg.ImportPath+" "+t.Name, lsp.SKMethod, fs, v.Decl.Name.NamePos))
				}
				for _, v := range t.Consts {
					for _, name := range v.Names {
						results.Collect(toSym(name, buildPkg.ImportPath, lsp.SKConstant, fs, v.Decl.TokPos))
					}
				}
				for _, v := range t.Vars {
					for _, name := range v.Names {
						results.Collect(toSym(name, buildPkg.ImportPath, lsp.SKField, fs, v.Decl.TokPos))
					}
				}
			}
			for _, v := range docPkg.Consts {
				for _, name := range v.Names {
					results.Collect(toSym(name, buildPkg.ImportPath, lsp.SKConstant, fs, v.Decl.TokPos))
				}
			}
			for _, v := range docPkg.Vars {
				for _, name := range v.Names {
					results.Collect(toSym(name, buildPkg.ImportPath, lsp.SKVariable, fs, v.Decl.TokPos))
				}
			}
			for _, v := range docPkg.Funcs {
				results.Collect(toSym(v.Name, buildPkg.ImportPath, lsp.SKFunction, fs, v.Decl.Name.NamePos))
			}
		}
	}
	sort.Sort(&results)
	if len(results.results) > params.Limit && params.Limit > 0 {
		results.results = results.results[:params.Limit]
	}
	return results.Results(), nil
}

// parseDir mirrors parser.ParseDir, but uses the passed in build context's VFS. In other words,
// buildutil.parseFile : parser.ParseFile :: parseDir : parser.ParseDir
func parseDir(fset *token.FileSet, bctx *build.Context, path string, filter func(os.FileInfo) bool, mode parser.Mode) (pkgs map[string]*ast.Package, first error) {
	list, err := buildutil.ReadDir(bctx, path)
	if err != nil {
		return nil, err
	}

	pkgs = make(map[string]*ast.Package)
	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".go") && (filter == nil || filter(d)) {
			filename := filepath.Join(path, d.Name())
			if src, err := buildutil.ParseFile(fset, bctx, nil, filepath.Join(path, d.Name()), filename, mode); err == nil {
				name := src.Name.Name
				pkg, found := pkgs[name]
				if !found {
					pkg = &ast.Package{
						Name:  name,
						Files: make(map[string]*ast.File),
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
