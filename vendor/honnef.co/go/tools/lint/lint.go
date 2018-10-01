// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lint provides the foundation for tools like gosimple.
package lint // import "honnef.co/go/tools/lint"

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/tools/go/loader"
	"honnef.co/go/tools/ssa"
	"honnef.co/go/tools/ssa/ssautil"
)

type Job struct {
	Program *Program

	checker  string
	check    string
	problems []Problem
}

type Ignore interface {
	Match(p Problem) bool
}

type LineIgnore struct {
	File    string
	Line    int
	Checks  []string
	matched bool
	pos     token.Pos
}

func (li *LineIgnore) Match(p Problem) bool {
	if p.Position.Filename != li.File || p.Position.Line != li.Line {
		return false
	}
	for _, c := range li.Checks {
		if m, _ := filepath.Match(c, p.Check); m {
			li.matched = true
			return true
		}
	}
	return false
}

func (li *LineIgnore) String() string {
	matched := "not matched"
	if li.matched {
		matched = "matched"
	}
	return fmt.Sprintf("%s:%d %s (%s)", li.File, li.Line, strings.Join(li.Checks, ", "), matched)
}

type FileIgnore struct {
	File   string
	Checks []string
}

func (fi *FileIgnore) Match(p Problem) bool {
	if p.Position.Filename != fi.File {
		return false
	}
	for _, c := range fi.Checks {
		if m, _ := filepath.Match(c, p.Check); m {
			return true
		}
	}
	return false
}

type GlobIgnore struct {
	Pattern string
	Checks  []string
}

func (gi *GlobIgnore) Match(p Problem) bool {
	if gi.Pattern != "*" {
		pkgpath := p.Package.Path()
		if strings.HasSuffix(pkgpath, "_test") {
			pkgpath = pkgpath[:len(pkgpath)-len("_test")]
		}
		name := filepath.Join(pkgpath, filepath.Base(p.Position.Filename))
		if m, _ := filepath.Match(gi.Pattern, name); !m {
			return false
		}
	}
	for _, c := range gi.Checks {
		if m, _ := filepath.Match(c, p.Check); m {
			return true
		}
	}
	return false
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
	pos      token.Pos
	Position token.Position // position in source file
	Text     string         // the prose that describes the problem
	Check    string
	Checker  string
	Package  *types.Package
	Ignored  bool
}

func (p *Problem) String() string {
	if p.Check == "" {
		return p.Text
	}
	return fmt.Sprintf("%s (%s)", p.Text, p.Check)
}

type Checker interface {
	Name() string
	Prefix() string
	Init(*Program)
	Funcs() map[string]Func
}

// A Linter lints Go source code.
type Linter struct {
	Checker       Checker
	Ignores       []Ignore
	GoVersion     int
	ReturnIgnored bool

	automaticIgnores []Ignore
}

func (l *Linter) ignore(p Problem) bool {
	ignored := false
	for _, ig := range l.automaticIgnores {
		// We cannot short-circuit these, as we want to record, for
		// each ignore, whether it matched or not.
		if ig.Match(p) {
			ignored = true
		}
	}
	if ignored {
		// no need to execute other ignores if we've already had a
		// match.
		return true
	}
	for _, ig := range l.Ignores {
		// We can short-circuit here, as we aren't tracking any
		// information.
		if ig.Match(p) {
			return true
		}
	}

	return false
}

func (prog *Program) File(node Positioner) *ast.File {
	return prog.tokenFileMap[prog.SSA.Fset.File(node.Pos())]
}

func (j *Job) File(node Positioner) *ast.File {
	return j.Program.File(node)
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
	pi, pj := ps.ps[i].Position, ps.ps[j].Position

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

func parseDirective(s string) (cmd string, args []string) {
	if !strings.HasPrefix(s, "//lint:") {
		return "", nil
	}
	s = strings.TrimPrefix(s, "//lint:")
	fields := strings.Split(s, " ")
	return fields[0], fields[1:]
}

func (l *Linter) Lint(lprog *loader.Program, conf *loader.Config) []Problem {
	ssaprog := ssautil.CreateProgram(lprog, ssa.GlobalDebug)
	ssaprog.Build()
	pkgMap := map[*ssa.Package]*Pkg{}
	var pkgs []*Pkg
	for _, pkginfo := range lprog.InitialPackages() {
		ssapkg := ssaprog.Package(pkginfo.Pkg)
		var bp *build.Package
		if len(pkginfo.Files) != 0 {
			path := lprog.Fset.Position(pkginfo.Files[0].Pos()).Filename
			dir := filepath.Dir(path)
			var err error
			ctx := conf.Build
			if ctx == nil {
				ctx = &build.Default
			}
			bp, err = ctx.ImportDir(dir, 0)
			if err != nil {
				// shouldn't happen
			}
		}
		pkg := &Pkg{
			Package:  ssapkg,
			Info:     pkginfo,
			BuildPkg: bp,
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
			prog.astFileMap[f] = pkgMap[ssapkg]
		}
	}

	for _, pkginfo := range lprog.AllPackages {
		for _, f := range pkginfo.Files {
			tf := lprog.Fset.File(f.Pos())
			prog.tokenFileMap[tf] = f
		}
	}

	var out []Problem
	l.automaticIgnores = nil
	for _, pkginfo := range lprog.InitialPackages() {
		for _, f := range pkginfo.Files {
			cm := ast.NewCommentMap(lprog.Fset, f, f.Comments)
			for node, cgs := range cm {
				for _, cg := range cgs {
					for _, c := range cg.List {
						if !strings.HasPrefix(c.Text, "//lint:") {
							continue
						}
						cmd, args := parseDirective(c.Text)
						switch cmd {
						case "ignore", "file-ignore":
							if len(args) < 2 {
								// FIXME(dh): this causes duplicated warnings when using megacheck
								p := Problem{
									pos:      c.Pos(),
									Position: prog.DisplayPosition(c.Pos()),
									Text:     "malformed linter directive; missing the required reason field?",
									Check:    "",
									Checker:  l.Checker.Name(),
									Package:  nil,
								}
								out = append(out, p)
								continue
							}
						default:
							// unknown directive, ignore
							continue
						}
						checks := strings.Split(args[0], ",")
						pos := prog.DisplayPosition(node.Pos())
						var ig Ignore
						switch cmd {
						case "ignore":
							ig = &LineIgnore{
								File:   pos.Filename,
								Line:   pos.Line,
								Checks: checks,
								pos:    c.Pos(),
							}
						case "file-ignore":
							ig = &FileIgnore{
								File:   pos.Filename,
								Checks: checks,
							}
						}
						l.automaticIgnores = append(l.automaticIgnores, ig)
					}
				}
			}
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
			checker: l.Checker.Name(),
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

	for _, j := range jobs {
		for _, p := range j.problems {
			p.Ignored = l.ignore(p)
			if l.ReturnIgnored || !p.Ignored {
				out = append(out, p)
			}
		}
	}

	for _, ig := range l.automaticIgnores {
		ig, ok := ig.(*LineIgnore)
		if !ok {
			continue
		}
		if ig.matched {
			continue
		}
		for _, c := range ig.Checks {
			idx := strings.IndexFunc(c, func(r rune) bool {
				return unicode.IsNumber(r)
			})
			if idx == -1 {
				// malformed check name, backing out
				continue
			}
			if c[:idx] != l.Checker.Prefix() {
				// not for this checker
				continue
			}
			p := Problem{
				pos:      ig.pos,
				Position: prog.DisplayPosition(ig.pos),
				Text:     "this linter directive didn't match anything; should it be removed?",
				Check:    "",
				Checker:  l.Checker.Name(),
				Package:  nil,
			}
			out = append(out, p)
		}
	}

	sort.Sort(byPosition{lprog.Fset, out})
	return out
}

// Pkg represents a package being linted.
type Pkg struct {
	*ssa.Package
	Info     *loader.PackageInfo
	BuildPkg *build.Package
}

type Positioner interface {
	Pos() token.Pos
}

func (prog *Program) DisplayPosition(p token.Pos) token.Position {
	// The //line compiler directive can be used to change the file
	// name and line numbers associated with code. This can, for
	// example, be used by code generation tools. The most prominent
	// example is 'go tool cgo', which uses //line directives to refer
	// back to the original source code.
	//
	// In the context of our linters, we need to treat these
	// directives differently depending on context. For cgo files, we
	// want to honour the directives, so that line numbers are
	// adjusted correctly. For all other files, we want to ignore the
	// directives, so that problems are reported at their actual
	// position and not, for example, a yacc grammar file. This also
	// affects the ignore mechanism, since it operates on the position
	// information stored within problems. With this implementation, a
	// user will ignore foo.go, not foo.y

	pkg := prog.astFileMap[prog.tokenFileMap[prog.Prog.Fset.File(p)]]
	bp := pkg.BuildPkg
	adjPos := prog.Prog.Fset.Position(p)
	if bp == nil {
		// couldn't find the package for some reason (deleted? faulty
		// file system?)
		return adjPos
	}
	base := filepath.Base(adjPos.Filename)
	for _, f := range bp.CgoFiles {
		if f == base {
			// this is a cgo file, use the adjusted position
			return adjPos
		}
	}
	// not a cgo file, ignore //line directives
	return prog.Prog.Fset.PositionFor(p, false)
}

func (j *Job) Errorf(n Positioner, format string, args ...interface{}) *Problem {
	tf := j.Program.SSA.Fset.File(n.Pos())
	f := j.Program.tokenFileMap[tf]
	pkg := j.Program.astFileMap[f].Pkg

	pos := j.Program.DisplayPosition(n.Pos())
	problem := Problem{
		pos:      n.Pos(),
		Position: pos,
		Text:     fmt.Sprintf(format, args...),
		Check:    j.check,
		Checker:  j.checker,
		Package:  pkg,
	}
	j.problems = append(j.problems, problem)
	return &j.problems[len(j.problems)-1]
}

func (j *Job) NodePackage(node Positioner) *Pkg {
	f := j.File(node)
	return j.Program.astFileMap[f]
}
