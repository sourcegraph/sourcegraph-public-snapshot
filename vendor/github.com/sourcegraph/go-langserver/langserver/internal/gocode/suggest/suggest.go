package suggest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/go-langserver/langserver/internal/gocode/lookdot"
)

type Config struct {
	Importer types.Importer
	Logf     func(fmt string, args ...interface{})
	Builtin  bool
}

type packageAnalysis struct {
	fset *token.FileSet
	pos  token.Pos
	pkg  *types.Package
}

// Suggest returns a list of suggestion candidates and the length of
// the text that should be replaced, if any.
func (c *Config) Suggest(filename string, data []byte, cursor int) ([]Candidate, int, error) {
	if cursor < 0 {
		return nil, 0, nil
	}

	a, err := c.analyzePackage(filename, data, cursor)
	if err != nil {
		return nil, 0, err
	}
	fset := a.fset
	pos := a.pos
	pkg := a.pkg
	if pkg == nil {
		return nil, 0, nil
	}
	scope := pkg.Scope().Innermost(pos)

	ctx, expr, partial := deduceCursorContext(data, cursor)
	b := candidateCollector{
		localpkg: pkg,
		partial:  partial,
		filter:   objectFilters[partial],
		builtin:  ctx != selectContext && c.Builtin,
	}

	switch ctx {
	case selectContext:
		tv, _ := types.Eval(fset, pkg, pos, expr)
		if lookdot.Walk(&tv, b.appendObject) {
			break
		}

		_, obj := scope.LookupParent(expr, pos)
		if pkgName, isPkg := obj.(*types.PkgName); isPkg {
			c.packageCandidates(pkgName.Imported(), &b)
			break
		}

		return nil, 0, nil

	case compositeLiteralContext:
		tv, _ := types.Eval(fset, pkg, pos, expr)
		if tv.IsType() {
			if _, isStruct := tv.Type.Underlying().(*types.Struct); isStruct {
				c.fieldNameCandidates(tv.Type, &b)
				break
			}
		}

		fallthrough
	default:
		c.scopeCandidates(scope, pos, &b)
	}

	res := b.getCandidates()
	if len(res) == 0 {
		return nil, 0, nil
	}
	return res, len(partial), nil
}

func (c *Config) analyzePackage(filename string, data []byte, cursor int) (*packageAnalysis, error) {
	// If we're in trailing white space at the end of a scope,
	// sometimes go/types doesn't recognize that variables should
	// still be in scope there.
	filesemi := bytes.Join([][]byte{data[:cursor], []byte(";"), data[cursor:]}, nil)

	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, filename, filesemi, parser.AllErrors)
	if err != nil {
		c.logParseError("Error parsing input file (outer block)", err)
	}
	astPos := fileAST.Pos()
	if astPos == 0 {
		return &packageAnalysis{fset: nil, pos: token.NoPos, pkg: nil}, nil
	}
	pos := fset.File(astPos).Pos(cursor)

	files := []*ast.File{fileAST}
	otherPkgFiles, err := c.findOtherPackageFiles(filename, fileAST.Name.Name)
	if err != nil {
		return nil, err
	}
	for _, otherName := range otherPkgFiles {
		ast, err := parser.ParseFile(fset, otherName, nil, 0)
		if err != nil {
			c.logParseError("Error parsing other file", err)
		}
		files = append(files, ast)
	}

	// Clear any function bodies other than where the cursor
	// is. They're not relevant to suggestions and only slow down
	// typechecking.
	for _, file := range files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && (pos < fd.Pos() || pos >= fd.End()) {
				fd.Body = nil
			}
		}
	}

	cfg := types.Config{
		Importer: c.Importer,
		Error:    func(err error) {},
	}
	pkg, _ := cfg.Check("", fset, files, nil)

	return &packageAnalysis{fset: fset, pos: pos, pkg: pkg}, nil
}

func (c *Config) fieldNameCandidates(typ types.Type, b *candidateCollector) {
	s := typ.Underlying().(*types.Struct)
	for i, n := 0, s.NumFields(); i < n; i++ {
		b.appendObject(s.Field(i))
	}
}

func (c *Config) packageCandidates(pkg *types.Package, b *candidateCollector) {
	c.scopeCandidates(pkg.Scope(), token.NoPos, b)
}

func (c *Config) scopeCandidates(scope *types.Scope, pos token.Pos, b *candidateCollector) {
	seen := make(map[string]bool)
	for scope != nil {
		for _, name := range scope.Names() {
			if seen[name] {
				continue
			}
			seen[name] = true
			_, obj := scope.LookupParent(name, pos)
			if obj != nil {
				b.appendObject(obj)
			}
		}
		scope = scope.Parent()
	}
}

func (c *Config) logParseError(intro string, err error) {
	if c.Logf == nil {
		return
	}
	if el, ok := err.(scanner.ErrorList); ok {
		c.Logf("%s:", intro)
		for _, er := range el {
			c.Logf(" %s", er)
		}
	} else {
		c.Logf("%s: %s", intro, err)
	}
}

func (c *Config) findOtherPackageFiles(filename, pkgName string) ([]string, error) {
	if filename == "" {
		return nil, nil
	}

	dir, file := filepath.Split(filename)
	dents, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read dir: %v", err)
	}
	isTestFile := strings.HasSuffix(file, "_test.go")

	// TODO(mdempsky): Use go/build.(*Context).MatchFile or
	// something to properly handle build tags?
	var out []string
	for _, dent := range dents {
		name := dent.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}
		if name == file || !strings.HasSuffix(name, ".go") {
			continue
		}
		if !isTestFile && strings.HasSuffix(name, "_test.go") {
			continue
		}

		abspath := filepath.Join(dir, name)
		if pkgNameFor(abspath) == pkgName {
			out = append(out, abspath)
		}
	}

	return out, nil
}

func pkgNameFor(filename string) string {
	file, _ := parser.ParseFile(token.NewFileSet(), filename, nil, parser.PackageClauseOnly)
	return file.Name.Name
}
