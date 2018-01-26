// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lint provides the foundation for tools like gosimple.
package lint // import "honnef.co/go/tools/lint"

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/printer"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
	"honnef.co/go/tools/ssa"
	"honnef.co/go/tools/ssa/ssautil"
)

type Job struct {
	Program *Program

	check    string
	problems []Problem
}

type Ignore struct {
	Pattern string
	Checks  []string
}

type Program struct {
	SSA  *ssa.Program
	Prog *loader.Program
	// TODO(dh): Rename to InitialPackages?
	Packages         []*Pkg
	InitialFunctions []*ssa.Function
	AllFunctions     []*ssa.Function
	Files            []*ast.File
	Info             *types.Info
	GoVersion        int

	tokenFileMap map[*token.File]*ast.File
	astFileMap   map[*ast.File]*Pkg
}

type Func func(*Job)

// Problem represents a problem in some source code.
type Problem struct {
	Position token.Pos // position in source file
	Text     string    // the prose that describes the problem
}

func (p *Problem) String() string {
	return p.Text
}

type Checker interface {
	Init(*Program)
	Funcs() map[string]Func
}

// A Linter lints Go source code.
type Linter struct {
	Checker   Checker
	Ignores   []Ignore
	GoVersion int
}

func (l *Linter) ignore(j *Job, p Problem) bool {
	tf := j.Program.SSA.Fset.File(p.Position)
	f := j.Program.tokenFileMap[tf]
	pkg := j.Program.astFileMap[f].Pkg

	for _, ig := range l.Ignores {
		pkgpath := pkg.Path()
		if strings.HasSuffix(pkgpath, "_test") {
			pkgpath = pkgpath[:len(pkgpath)-len("_test")]
		}
		name := filepath.Join(pkgpath, filepath.Base(tf.Name()))
		if m, _ := filepath.Match(ig.Pattern, name); !m {
			continue
		}
		for _, c := range ig.Checks {
			if m, _ := filepath.Match(c, j.check); m {
				return true
			}
		}
	}
	return false
}

func (j *Job) File(node Positioner) *ast.File {
	return j.Program.tokenFileMap[j.Program.SSA.Fset.File(node.Pos())]
}

// TODO(dh): switch to sort.Slice when Go 1.9 lands.
type byPosition struct {
	fset *token.FileSet
	ps   []Problem
}

func (ps byPosition) Len() int {
	return len(ps.ps)
}

func (ps byPosition) Less(i int, j int) bool {
	pi, pj := ps.fset.Position(ps.ps[i].Position), ps.fset.Position(ps.ps[j].Position)

	if pi.Filename != pj.Filename {
		return pi.Filename < pj.Filename
	}
	if pi.Line != pj.Line {
		return pi.Line < pj.Line
	}
	if pi.Column != pj.Column {
		return pi.Column < pj.Column
	}

	return ps.ps[i].Text < ps.ps[j].Text
}

func (ps byPosition) Swap(i int, j int) {
	ps.ps[i], ps.ps[j] = ps.ps[j], ps.ps[i]
}

func (l *Linter) Lint(lprog *loader.Program) []Problem {
	ssaprog := ssautil.CreateProgram(lprog, ssa.GlobalDebug)
	ssaprog.Build()
	pkgMap := map[*ssa.Package]*Pkg{}
	var pkgs []*Pkg
	for _, pkginfo := range lprog.InitialPackages() {
		ssapkg := ssaprog.Package(pkginfo.Pkg)
		pkg := &Pkg{
			Package: ssapkg,
			Info:    pkginfo,
		}
		pkgMap[ssapkg] = pkg
		pkgs = append(pkgs, pkg)
	}
	prog := &Program{
		SSA:          ssaprog,
		Prog:         lprog,
		Packages:     pkgs,
		Info:         &types.Info{},
		GoVersion:    l.GoVersion,
		tokenFileMap: map[*token.File]*ast.File{},
		astFileMap:   map[*ast.File]*Pkg{},
	}
	initial := map[*types.Package]struct{}{}
	for _, pkg := range pkgs {
		initial[pkg.Info.Pkg] = struct{}{}
	}
	for fn := range ssautil.AllFunctions(ssaprog) {
		if fn.Pkg == nil {
			continue
		}
		prog.AllFunctions = append(prog.AllFunctions, fn)
		if _, ok := initial[fn.Pkg.Pkg]; ok {
			prog.InitialFunctions = append(prog.InitialFunctions, fn)
		}
	}
	for _, pkg := range pkgs {
		prog.Files = append(prog.Files, pkg.Info.Files...)

		ssapkg := ssaprog.Package(pkg.Info.Pkg)
		for _, f := range pkg.Info.Files {
			tf := lprog.Fset.File(f.Pos())
			prog.tokenFileMap[tf] = f
			prog.astFileMap[f] = pkgMap[ssapkg]
		}
	}

	sizes := struct {
		types      int
		defs       int
		uses       int
		implicits  int
		selections int
		scopes     int
	}{}
	for _, pkg := range pkgs {
		sizes.types += len(pkg.Info.Info.Types)
		sizes.defs += len(pkg.Info.Info.Defs)
		sizes.uses += len(pkg.Info.Info.Uses)
		sizes.implicits += len(pkg.Info.Info.Implicits)
		sizes.selections += len(pkg.Info.Info.Selections)
		sizes.scopes += len(pkg.Info.Info.Scopes)
	}
	prog.Info.Types = make(map[ast.Expr]types.TypeAndValue, sizes.types)
	prog.Info.Defs = make(map[*ast.Ident]types.Object, sizes.defs)
	prog.Info.Uses = make(map[*ast.Ident]types.Object, sizes.uses)
	prog.Info.Implicits = make(map[ast.Node]types.Object, sizes.implicits)
	prog.Info.Selections = make(map[*ast.SelectorExpr]*types.Selection, sizes.selections)
	prog.Info.Scopes = make(map[ast.Node]*types.Scope, sizes.scopes)
	for _, pkg := range pkgs {
		for k, v := range pkg.Info.Info.Types {
			prog.Info.Types[k] = v
		}
		for k, v := range pkg.Info.Info.Defs {
			prog.Info.Defs[k] = v
		}
		for k, v := range pkg.Info.Info.Uses {
			prog.Info.Uses[k] = v
		}
		for k, v := range pkg.Info.Info.Implicits {
			prog.Info.Implicits[k] = v
		}
		for k, v := range pkg.Info.Info.Selections {
			prog.Info.Selections[k] = v
		}
		for k, v := range pkg.Info.Info.Scopes {
			prog.Info.Scopes[k] = v
		}
	}
	l.Checker.Init(prog)

	funcs := l.Checker.Funcs()
	var keys []string
	for k := range funcs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var jobs []*Job
	for _, k := range keys {
		j := &Job{
			Program: prog,
			check:   k,
		}
		jobs = append(jobs, j)
	}
	wg := &sync.WaitGroup{}
	for _, j := range jobs {
		wg.Add(1)
		go func(j *Job) {
			defer wg.Done()
			fn := funcs[j.check]
			if fn == nil {
				return
			}
			fn(j)
		}(j)
	}
	wg.Wait()

	var out []Problem
	for _, j := range jobs {
		for _, p := range j.problems {
			if !l.ignore(j, p) {
				out = append(out, p)
			}
		}
	}

	sort.Sort(byPosition{lprog.Fset, out})
	return out
}

// Pkg represents a package being linted.
type Pkg struct {
	*ssa.Package
	Info *loader.PackageInfo
}

type packager interface {
	Package() *ssa.Package
}

func IsExample(fn *ssa.Function) bool {
	if !strings.HasPrefix(fn.Name(), "Example") {
		return false
	}
	f := fn.Prog.Fset.File(fn.Pos())
	if f == nil {
		return false
	}
	return strings.HasSuffix(f.Name(), "_test.go")
}

func (j *Job) IsInTest(node Positioner) bool {
	f := j.Program.SSA.Fset.File(node.Pos())
	return f != nil && strings.HasSuffix(f.Name(), "_test.go")
}

func (j *Job) IsInMain(node Positioner) bool {
	if node, ok := node.(packager); ok {
		return node.Package().Pkg.Name() == "main"
	}
	pkg := j.NodePackage(node)
	if pkg == nil {
		return false
	}
	return pkg.Pkg.Name() == "main"
}

type Positioner interface {
	Pos() token.Pos
}

func (j *Job) Errorf(n Positioner, format string, args ...interface{}) *Problem {
	problem := Problem{
		Position: n.Pos(),
		Text:     fmt.Sprintf(format, args...) + fmt.Sprintf(" (%s)", j.check),
	}
	j.problems = append(j.problems, problem)
	return &j.problems[len(j.problems)-1]
}

func (j *Job) Render(x interface{}) string {
	fset := j.Program.SSA.Fset
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}

func (j *Job) RenderArgs(args []ast.Expr) string {
	var ss []string
	for _, arg := range args {
		ss = append(ss, j.Render(arg))
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

func IsZero(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "0"
}

func (j *Job) IsNil(expr ast.Expr) bool {
	return j.Program.Info.Types[expr].IsNil()
}

func (j *Job) BoolConst(expr ast.Expr) bool {
	val := j.Program.Info.ObjectOf(expr.(*ast.Ident)).(*types.Const).Val()
	return constant.BoolVal(val)
}

func (j *Job) IsBoolConst(expr ast.Expr) bool {
	// We explicitly don't support typed bools because more often than
	// not, custom bool types are used as binary enums and the
	// explicit comparison is desired.

	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := j.Program.Info.ObjectOf(ident)
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

func (j *Job) ExprToInt(expr ast.Expr) (int64, bool) {
	tv := j.Program.Info.Types[expr]
	if tv.Value == nil {
		return 0, false
	}
	if tv.Value.Kind() != constant.Int {
		return 0, false
	}
	return constant.Int64Val(tv.Value)
}

func (j *Job) ExprToString(expr ast.Expr) (string, bool) {
	val := j.Program.Info.Types[expr].Value
	if val == nil {
		return "", false
	}
	if val.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(val), true
}

func (j *Job) NodePackage(node Positioner) *Pkg {
	f := j.File(node)
	return j.Program.astFileMap[f]
}

func IsGenerated(f *ast.File) bool {
	comments := f.Comments
	if len(comments) > 0 {
		comment := comments[0].Text()
		return strings.Contains(comment, "Code generated by") ||
			strings.Contains(comment, "DO NOT EDIT")
	}
	return false
}

func Preamble(f *ast.File) string {
	cutoff := f.Package
	if f.Doc != nil {
		cutoff = f.Doc.Pos()
	}
	var out []string
	for _, cmt := range f.Comments {
		if cmt.Pos() >= cutoff {
			break
		}
		out = append(out, cmt.Text())
	}
	return strings.Join(out, "\n")
}

func IsPointerLike(T types.Type) bool {
	switch T.Underlying().(type) {
	case *types.Interface, *types.Chan, *types.Map, *types.Pointer:
		return true
	}
	return false
}

func (j *Job) IsGoVersion(minor int) bool {
	return j.Program.GoVersion >= minor
}

func (j *Job) IsCallToAST(node ast.Node, name string) bool {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	fn, ok := j.Program.Info.ObjectOf(sel.Sel).(*types.Func)
	return ok && fn.FullName() == name
}

func (j *Job) IsCallToAnyAST(node ast.Node, names ...string) bool {
	for _, name := range names {
		if j.IsCallToAST(node, name) {
			return true
		}
	}
	return false
}

func CallName(call *ssa.CallCommon) string {
	if call.IsInvoke() {
		return ""
	}
	switch v := call.Value.(type) {
	case *ssa.Function:
		fn, ok := v.Object().(*types.Func)
		if !ok {
			return ""
		}
		return fn.FullName()
	case *ssa.Builtin:
		return v.Name()
	}
	return ""
}

func IsCallTo(call *ssa.CallCommon, name string) bool {
	return CallName(call) == name
}

func FilterDebug(instr []ssa.Instruction) []ssa.Instruction {
	var out []ssa.Instruction
	for _, ins := range instr {
		if _, ok := ins.(*ssa.DebugRef); !ok {
			out = append(out, ins)
		}
	}
	return out
}

func NodeFns(pkgs []*Pkg) map[ast.Node]*ssa.Function {
	out := map[ast.Node]*ssa.Function{}

	wg := &sync.WaitGroup{}
	chNodeFns := make(chan map[ast.Node]*ssa.Function, runtime.NumCPU()*2)
	for _, pkg := range pkgs {
		pkg := pkg
		wg.Add(1)
		go func() {
			m := map[ast.Node]*ssa.Function{}
			for _, f := range pkg.Info.Files {
				ast.Walk(&globalVisitor{m, pkg, f}, f)
			}
			chNodeFns <- m
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(chNodeFns)
	}()

	for nodeFns := range chNodeFns {
		for k, v := range nodeFns {
			out[k] = v
		}
	}

	return out
}

type globalVisitor struct {
	m   map[ast.Node]*ssa.Function
	pkg *Pkg
	f   *ast.File
}

func (v *globalVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.CallExpr:
		v.m[node] = v.pkg.Func("init")
		return v
	case *ast.FuncDecl, *ast.FuncLit:
		nv := &fnVisitor{v.m, v.f, v.pkg, nil}
		return nv.Visit(node)
	default:
		return v
	}
}

type fnVisitor struct {
	m     map[ast.Node]*ssa.Function
	f     *ast.File
	pkg   *Pkg
	ssafn *ssa.Function
}

func (v *fnVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		var ssafn *ssa.Function
		ssafn = v.pkg.Prog.FuncValue(v.pkg.Info.ObjectOf(node.Name).(*types.Func))
		v.m[node] = ssafn
		if ssafn == nil {
			return nil
		}
		return &fnVisitor{v.m, v.f, v.pkg, ssafn}
	case *ast.FuncLit:
		var ssafn *ssa.Function
		path, _ := astutil.PathEnclosingInterval(v.f, node.Pos(), node.Pos())
		ssafn = ssa.EnclosingFunction(v.pkg.Package, path)
		v.m[node] = ssafn
		if ssafn == nil {
			return nil
		}
		return &fnVisitor{v.m, v.f, v.pkg, ssafn}
	case nil:
		return nil
	default:
		v.m[node] = v.ssafn
		return v
	}
}
