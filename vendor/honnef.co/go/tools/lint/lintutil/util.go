// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lintutil provides helpers for writing linter command lines.
package lintutil // import "honnef.co/go/tools/lint/lintutil"

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"honnef.co/go/tools/lint"
	"honnef.co/go/tools/version"

	"github.com/kisielk/gotool"
	"golang.org/x/tools/go/loader"
)

func usage(name string, flags *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", name)
		fmt.Fprintf(os.Stderr, "\t%s [flags] # runs on package in current directory\n", name)
		fmt.Fprintf(os.Stderr, "\t%s [flags] packages\n", name)
		fmt.Fprintf(os.Stderr, "\t%s [flags] directory\n", name)
		fmt.Fprintf(os.Stderr, "\t%s [flags] files... # must be a single package\n", name)
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
	}
}

type runner struct {
	checker lint.Checker
	tags    []string
	ignores []lint.Ignore
	version int
}

func (runner runner) resolveRelative(importPaths []string) (goFiles bool, err error) {
	if len(importPaths) == 0 {
		return false, nil
	}
	if strings.HasSuffix(importPaths[0], ".go") {
		// User is specifying a package in terms of .go files, don't resolve
		return true, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	ctx := build.Default
	ctx.BuildTags = runner.tags
	for i, path := range importPaths {
		bpkg, err := ctx.Import(path, wd, build.FindOnly)
		if err != nil {
			return false, fmt.Errorf("can't load package %q: %v", path, err)
		}
		importPaths[i] = bpkg.ImportPath
	}
	return false, nil
}

func parseIgnore(s string) ([]lint.Ignore, error) {
	var out []lint.Ignore
	if len(s) == 0 {
		return nil, nil
	}
	for _, part := range strings.Fields(s) {
		p := strings.Split(part, ":")
		if len(p) != 2 {
			return nil, errors.New("malformed ignore string")
		}
		path := p[0]
		checks := strings.Split(p[1], ",")
		out = append(out, lint.Ignore{Pattern: path, Checks: checks})
	}
	return out, nil
}

type versionFlag int

func (v *versionFlag) String() string {
	return fmt.Sprintf("1.%d", *v)
}

func (v *versionFlag) Set(s string) error {
	if len(s) < 3 {
		return errors.New("invalid Go version")
	}
	if s[0] != '1' {
		return errors.New("invalid Go version")
	}
	if s[1] != '.' {
		return errors.New("invalid Go version")
	}
	i, err := strconv.Atoi(s[2:])
	*v = versionFlag(i)
	return err
}

func (v *versionFlag) Get() interface{} {
	return int(*v)
}

func FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = usage(name, flags)
	flags.Float64("min_confidence", 0, "Deprecated; use -ignore instead")
	flags.String("tags", "", "List of `build tags`")
	flags.String("ignore", "", "Space separated list of checks to ignore, in the following format: 'import/path/file.go:Check1,Check2,...' Both the import path and file name sections support globbing, e.g. 'os/exec/*_test.go'")
	flags.Bool("tests", true, "Include tests")
	flags.Bool("version", false, "Print version and exit")

	tags := build.Default.ReleaseTags
	v := tags[len(tags)-1][2:]
	version := new(versionFlag)
	if err := version.Set(v); err != nil {
		panic(fmt.Sprintf("internal error: %s", err))
	}

	flags.Var(version, "go", "Target Go `version` in the format '1.x'")
	return flags
}

func ProcessFlagSet(c lint.Checker, fs *flag.FlagSet) {
	tags := fs.Lookup("tags").Value.(flag.Getter).Get().(string)
	ignore := fs.Lookup("ignore").Value.(flag.Getter).Get().(string)
	tests := fs.Lookup("tests").Value.(flag.Getter).Get().(bool)
	goVersion := fs.Lookup("go").Value.(flag.Getter).Get().(int)
	printVersion := fs.Lookup("version").Value.(flag.Getter).Get().(bool)

	if printVersion {
		version.Print()
		os.Exit(0)
	}

	ps, lprog, err := Lint(c, fs.Args(), &Options{
		Tags:      strings.Fields(tags),
		LintTests: tests,
		Ignores:   ignore,
		GoVersion: goVersion,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	unclean := false
	for _, p := range ps {
		unclean = true
		pos := lprog.Fset.Position(p.Position)
		fmt.Printf("%v: %s\n", relativePositionString(pos), p.Text)
	}
	if unclean {
		os.Exit(1)
	}
}

type Options struct {
	Tags      []string
	LintTests bool
	Ignores   string
	GoVersion int
}

func Lint(c lint.Checker, pkgs []string, opt *Options) ([]lint.Problem, *loader.Program, error) {
	// TODO(dh): Instead of returning the loader.Program, we should
	// store token.Position instead of token.Pos in lint.Problem.
	if opt == nil {
		opt = &Options{}
	}
	ignores, err := parseIgnore(opt.Ignores)
	if err != nil {
		return nil, nil, err
	}
	runner := &runner{
		checker: c,
		tags:    opt.Tags,
		ignores: ignores,
		version: opt.GoVersion,
	}
	paths := gotool.ImportPaths(pkgs)
	goFiles, err := runner.resolveRelative(paths)
	if err != nil {
		return nil, nil, err
	}
	ctx := build.Default
	ctx.BuildTags = runner.tags
	conf := &loader.Config{
		Build:      &ctx,
		ParserMode: parser.ParseComments,
		ImportPkgs: map[string]bool{},
	}
	if goFiles {
		conf.CreateFromFilenames("adhoc", paths...)
	} else {
		for _, path := range paths {
			conf.ImportPkgs[path] = opt.LintTests
		}
	}
	lprog, err := conf.Load()
	if err != nil {
		return nil, nil, err
	}
	return runner.lint(lprog), lprog, nil
}

func shortPath(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	if rel, err := filepath.Rel(cwd, path); err == nil && len(rel) < len(path) {
		return rel
	}
	return path
}

func relativePositionString(pos token.Position) string {
	s := shortPath(pos.Filename)
	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

func ProcessArgs(name string, c lint.Checker, args []string) {
	flags := FlagSet(name)
	flags.Parse(args)

	ProcessFlagSet(c, flags)
}

func (runner *runner) lint(lprog *loader.Program) []lint.Problem {
	l := &lint.Linter{
		Checker:   runner.checker,
		Ignores:   runner.ignores,
		GoVersion: runner.version,
	}
	return l.Lint(lprog)
}
