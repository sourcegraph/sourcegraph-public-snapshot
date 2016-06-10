// +build ignore

// doccover prints documentation coverage statistics and problems.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
)

var (
	printAll = flag.Bool("all", false, "print names of all undocumented definitions, not just packages and percentages")
)

var fset = token.NewFileSet()

func main() {
	log.SetFlags(0)
	flag.Parse()

	pkgs, err := goListPkgs(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(pkgs)

	var overall struct {
		Pkg             int
		PkgUndocumented int

		Total        int
		Undocumented int
	}

	for _, path := range pkgs {
		bpkg, err := build.Import(path, "", 0)
		if err != nil {
			log.Fatal(err)
		}

		pkgs, err := parser.ParseDir(fset, bpkg.Dir, mainGoFiles, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		for pkgName, pkg := range pkgs {
			if pkgName == "main" && len(pkgs) > 1 {
				// Skip, e.g., build-tag-disabled codegen files.
				continue
			}

			cov := docCoverage(doc.New(pkg, path, 0))
			fmt.Println(cov)

			overall.Pkg++
			if cov.PkgUndocumented {
				overall.PkgUndocumented++
			}
			overall.Total += cov.Total
			overall.Undocumented += len(cov.Undocumented)
		}
	}

	log.Println("# packages", pct(overall.Pkg-overall.PkgUndocumented, overall.Pkg))
	log.Println("# all defs", pct(overall.Total-overall.Undocumented, overall.Total))
}

func pct(n, d int) string {
	return fmt.Sprintf("%.1f%% (%d/%d)", 100*float64(n)/float64(d), n, d)
}

func mainGoFiles(fi os.FileInfo) bool {
	return !strings.HasSuffix(fi.Name(), "_test.go")
}

func goListPkgs(args []string) ([]string, error) {
	cmd := exec.Command("go", "list", "-e")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exec %v failed: %s", cmd.Args, err)
	}
	return strings.Split(string(bytes.TrimSpace(out)), "\n"), nil
}

type coverage struct {
	Pkg             string   // package import path
	Total           int      // number of definitions
	Undocumented    []string // names of undocumented definitions
	PkgUndocumented bool     // whether there is no package doc comment
}

func (c *coverage) add(name, doc string) {
	c.Total++
	if doc == "" {
		c.Undocumented = append(c.Undocumented, name)
		if name == "package" {
			c.PkgUndocumented = true
		}
	}
}

func (c coverage) String() string {
	undoc := len(c.Undocumented)
	doc := c.Total - undoc

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%-.1f%%\t%d/%d\t%s", 100*float64(doc)/float64(c.Total), doc, c.Total, c.Pkg)
	if c.PkgUndocumented {
		fmt.Fprint(&buf, " (no package doc)")
	}
	fmt.Fprintln(&buf)

	if *printAll {
		for _, name := range c.Undocumented {
			fmt.Fprintln(&buf, "\t\t -", name)
		}
	}

	return strings.TrimSpace(buf.String())
}

func docCoverage(pkg *doc.Package) coverage {
	var cov coverage
	cov.Pkg = pkg.ImportPath

	cov.add("package", pkg.Doc)

	addConsts := func(consts []*doc.Value) {
		for _, c := range consts {
			cov.add(fmt.Sprintf("const %s", strings.Join(c.Names, " ")), c.Doc)
		}
	}
	addVars := func(vars []*doc.Value) {
		for _, v := range pkg.Vars {
			cov.add(fmt.Sprintf("var %s", strings.Join(v.Names, " ")), v.Doc)
		}
	}
	addFuncs := func(funcs []*doc.Func) {
		for _, f := range pkg.Funcs {
			var name string
			if f.Recv == "" {
				name = fmt.Sprintf("func %s", f.Name)
			} else {
				name = fmt.Sprintf("func (%s) %s", f.Recv, f.Name)
			}
			cov.add(name, f.Doc)
		}
	}

	addConsts(pkg.Consts)
	addVars(pkg.Vars)
	addFuncs(pkg.Funcs)
	for _, t := range pkg.Types {
		cov.add(fmt.Sprintf("type %s", t.Name), t.Doc)
		addConsts(t.Consts)
		addVars(t.Vars)
		addFuncs(t.Funcs)
		addFuncs(t.Methods)
	}

	return cov
}
