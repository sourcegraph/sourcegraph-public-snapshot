// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lint provides the foundation for tools like gosimple.
package lint // import "honnef.co/go/lint"

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
	"honnef.co/go/ssa"
)

type Ignore struct {
	Pattern string
	Checks  []string
}

type Program struct {
	Prog     *loader.Program
	Packages []*Pkg
}

type Func func(*File)

// Problem represents a problem in some source code.
type Problem struct {
	Position token.Position // position in source file
	Text     string         // the prose that describes the problem

	// If the problem has a suggested fix (the minority case),
	// ReplacementLine is a full replacement for the relevant line of the source file.
	ReplacementLine string
}

func (p *Problem) String() string {
	return p.Text
}

type ByPosition []Problem

func (p ByPosition) Len() int      { return len(p) }
func (p ByPosition) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p ByPosition) Less(i, j int) bool {
	pi, pj := p[i].Position, p[j].Position

	if pi.Filename != pj.Filename {
		return pi.Filename < pj.Filename
	}
	if pi.Line != pj.Line {
		return pi.Line < pj.Line
	}
	if pi.Column != pj.Column {
		return pi.Column < pj.Column
	}

	return p[i].Text < p[j].Text
}

type Checker interface {
	Init(*Program)
	Funcs() map[string]Func
}

// A Linter lints Go source code.
type Linter struct {
	Checker Checker
	Ignores []Ignore
}

func buildPackage(pkg *types.Package, files []*ast.File, info *types.Info, fset *token.FileSet, mode ssa.BuilderMode) *ssa.Package {
	prog := ssa.NewProgram(fset, mode)

	// Create SSA packages for all imports.
	// Order is not significant.
	created := make(map[*types.Package]bool)
	var createAll func(pkgs []*types.Package)
	createAll = func(pkgs []*types.Package) {
		for _, p := range pkgs {
			if !created[p] {
				created[p] = true
				prog.CreatePackage(p, nil, nil, true)
				createAll(p.Imports())
			}
		}
	}
	createAll(pkg.Imports())

	// Create and build the primary package.
	ssapkg := prog.CreatePackage(pkg, files, info, false)
	ssapkg.Build()
	return ssapkg
}

func (l *Linter) ignore(f *File, check string) bool {
	for _, ig := range l.Ignores {
		pkg := f.Pkg.TypesPkg.Path()
		if strings.HasSuffix(pkg, "_test") {
			pkg = pkg[:len(pkg)-len("_test")]
		}
		name := filepath.Join(pkg, filepath.Base(f.Filename))
		if m, _ := filepath.Match(ig.Pattern, name); !m {
			continue
		}
		for _, c := range ig.Checks {
			if m, _ := filepath.Match(c, check); m {
				return true
			}
		}
	}
	return false
}

func (l *Linter) Lint(lprog *loader.Program) map[string][]Problem {
	var pkgs []*Pkg
	for _, pkginfo := range lprog.InitialPackages() {
		ssapkg := buildPackage(pkginfo.Pkg, pkginfo.Files, &pkginfo.Info, lprog.Fset, ssa.GlobalDebug)
		pkg := &Pkg{
			TypesPkg:  pkginfo.Pkg,
			TypesInfo: pkginfo.Info,
			SSAPkg:    ssapkg,
			PkgInfo:   pkginfo,
		}
		pkgs = append(pkgs, pkg)
	}
	prog := &Program{
		Prog:     lprog,
		Packages: pkgs,
	}
	l.Checker.Init(prog)

	funcs := l.Checker.Funcs()
	var keys []string
	for k := range funcs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := map[string][]Problem{}
	type result struct {
		path     string
		problems []Problem
	}
	work := make(chan *Pkg, runtime.NumCPU())
	res := make(chan result, runtime.NumCPU())
	worker := func(wg *sync.WaitGroup, work chan *Pkg, res chan result) {
		for pkg := range work {
			for _, file := range pkg.PkgInfo.Files {
				path := lprog.Fset.Position(file.Pos()).Filename
				for _, k := range keys {
					f := &File{
						Pkg:      pkg,
						File:     file,
						Filename: path,
						Fset:     lprog.Fset,
						Program:  lprog,
						check:    k,
					}

					fn := funcs[k]
					if fn == nil {
						continue
					}
					if l.ignore(f, k) {
						continue
					}
					fn(f)
				}
			}
			sort.Sort(ByPosition(pkg.problems))
			res <- result{pkg.PkgInfo.Pkg.Path(), pkg.problems}
		}
		wg.Done()
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(wg, work, res)
	}
	go func() {
		wg.Wait()
		close(res)
	}()
	go func() {
		for _, pkg := range pkgs {
			work <- pkg
		}
		close(work)
	}()
	for r := range res {
		out[r.path] = r.problems
	}
	return out
}

func (f *File) Source() []byte {
	if f.src != nil {
		return f.src
	}
	path := f.Fset.Position(f.File.Pos()).Filename
	if path != "" {
		f.src, _ = ioutil.ReadFile(path)
	}
	return f.src
}

// pkg represents a package being linted.
type Pkg struct {
	TypesPkg  *types.Package
	TypesInfo types.Info
	SSAPkg    *ssa.Package
	PkgInfo   *loader.PackageInfo

	problems []Problem
}

// file represents a file being linted.
type File struct {
	Pkg      *Pkg
	File     *ast.File
	Filename string
	Fset     *token.FileSet
	Program  *loader.Program
	src      []byte
	check    string
}

func (f *File) IsTest() bool { return strings.HasSuffix(f.Filename, "_test.go") }

type Positioner interface {
	Pos() token.Pos
}

func (f *File) Errorf(n Positioner, format string, args ...interface{}) *Problem {
	fmt.Println(f.Program.Fset.Position(f.File.Pos()))
	pos := f.Fset.Position(n.Pos())
	return f.Pkg.errorfAt(pos, f.check, format, args...)
}

func (p *Pkg) errorfAt(pos token.Position, check string, format string, args ...interface{}) *Problem {
	problem := Problem{
		Position: pos,
	}

	problem.Text = fmt.Sprintf(format, args...) + fmt.Sprintf(" (%s)", check)
	p.problems = append(p.problems, problem)
	return &p.problems[len(p.problems)-1]
}

func (p *Pkg) IsNamedType(typ types.Type, importPath, name string) bool {
	n, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	tn := n.Obj()
	return tn != nil && tn.Pkg() != nil && tn.Pkg().Path() == importPath && tn.Name() == name
}

func (f *File) IsMain() bool {
	return f.File.Name.Name == "main"
}

// exportedType reports whether typ is an exported type.
// It is imprecise, and will err on the side of returning true,
// such as for composite types.
func ExportedType(typ types.Type) bool {
	switch T := typ.(type) {
	case *types.Named:
		// Builtin types have no package.
		return T.Obj().Pkg() == nil || T.Obj().Exported()
	case *types.Map:
		return ExportedType(T.Key()) && ExportedType(T.Elem())
	case interface {
		Elem() types.Type
	}: // array, slice, pointer, chan
		return ExportedType(T.Elem())
	}
	// Be conservative about other types, such as struct, interface, etc.
	return true
}

func ReceiverType(fn *ast.FuncDecl) string {
	switch e := fn.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return e.X.(*ast.Ident).Name
	}
	panic(fmt.Sprintf("unknown method receiver AST node type %T", fn.Recv.List[0].Type))
}

func (f *File) Walk(fn func(ast.Node) bool) {
	ast.Inspect(f.File, fn)
}

func (f *File) Render(x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, f.Fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}

func (f *File) RenderArgs(args []ast.Expr) string {
	var ss []string
	for _, arg := range args {
		ss = append(ss, f.Render(arg))
	}
	return strings.Join(ss, ", ")
}

func IsIdent(expr ast.Expr, ident string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == ident
}

// isBlank returns whether id is the blank identifier "_".
// If id == nil, the answer is false.
func IsBlank(id ast.Expr) bool {
	ident, ok := id.(*ast.Ident)
	return ok && ident.Name == "_"
}

func IsPkgDot(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && IsIdent(sel.X, pkg) && IsIdent(sel.Sel, name)
}

func IsZero(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "0"
}

func IsOne(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "1"
}

func IsNil(expr ast.Expr) bool {
	// FIXME(dominikh): use type information
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

var basicTypeKinds = map[types.BasicKind]string{
	types.UntypedBool:    "bool",
	types.UntypedInt:     "int",
	types.UntypedRune:    "rune",
	types.UntypedFloat:   "float64",
	types.UntypedComplex: "complex128",
	types.UntypedString:  "string",
}

// isUntypedConst reports whether expr is an untyped constant,
// and indicates what its default type is.
// scope may be nil.
func (f *File) IsUntypedConst(expr ast.Expr) (defType string, ok bool) {
	// Re-evaluate expr outside of its context to see if it's untyped.
	// (An expr evaluated within, for example, an assignment context will get the type of the LHS.)
	exprStr := f.Render(expr)
	tv, err := types.Eval(f.Fset, f.Pkg.TypesPkg, expr.Pos(), exprStr)
	if err != nil {
		return "", false
	}
	if b, ok := tv.Type.(*types.Basic); ok {
		if dt, ok := basicTypeKinds[b.Kind()]; ok {
			return dt, true
		}
	}

	return "", false
}

func (f *File) BoolConst(expr ast.Expr) bool {
	val := f.Pkg.TypesInfo.ObjectOf(expr.(*ast.Ident)).(*types.Const).Val()
	return constant.BoolVal(val)
}

func (f *File) IsBoolConst(expr ast.Expr) bool {
	// We explicitly don't support typed bools because more often than
	// not, custom bool types are used as binary enums and the
	// explicit comparison is desired.

	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := f.Pkg.TypesInfo.ObjectOf(ident)
	c, ok := obj.(*types.Const)
	if !ok {
		return false
	}
	basic, ok := c.Type().(*types.Basic)
	if !ok {
		return false
	}
	if basic.Kind() != types.UntypedBool && basic.Kind() != types.Bool {
		return false
	}
	return true
}

func ExprToInt(expr ast.Expr) (string, bool) {
	switch y := expr.(type) {
	case *ast.BasicLit:
		if y.Kind != token.INT {
			return "", false
		}
		return y.Value, true
	case *ast.UnaryExpr:
		if y.Op != token.SUB && y.Op != token.ADD {
			return "", false
		}
		x, ok := y.X.(*ast.BasicLit)
		if !ok {
			return "", false
		}
		if x.Kind != token.INT {
			return "", false
		}
		v := constant.MakeFromLiteral(x.Value, x.Kind, 0)
		return constant.UnaryOp(y.Op, v, 0).String(), true
	default:
		return "", false
	}
}

func (f *File) EnclosingSSAFunction(node Positioner) *ssa.Function {
	path, _ := astutil.PathEnclosingInterval(f.File, node.Pos(), node.Pos())
	return ssa.EnclosingFunction(f.Pkg.SSAPkg, path)
}
