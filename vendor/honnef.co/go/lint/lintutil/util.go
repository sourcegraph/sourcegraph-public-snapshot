// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lintutil provides helpers for writing linter command lines.
package lintutil // import "honnef.co/go/lint/lintutil"

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"log"
	"os"
	"strings"

	"honnef.co/go/lint"

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

	unclean bool
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

func ProcessArgs(name string, c lint.Checker, args []string) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = usage(name, flags)
	flags.Float64("min_confidence", 0, "Deprecated; use -ignore instead")
	tags := flags.String("tags", "", "List of `build tags`")
	ignore := flags.String("ignore", "", "Space separated list of checks to ignore, in the following format: 'import/path/file.go:Check1,Check2,...' Both the import path and file name sections support globbing, e.g. 'os/exec/*_test.go'")
	flags.Parse(args)

	ignores, err := parseIgnore(*ignore)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	runner := &runner{
		checker: c,
		tags:    strings.Fields(*tags),
		ignores: ignores,
	}
	paths := gotool.ImportPaths(flags.Args())
	goFiles, err := runner.resolveRelative(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		runner.unclean = true
	}
	ctx := build.Default
	ctx.BuildTags = runner.tags
	conf := &loader.Config{
		Build:      &ctx,
		ParserMode: parser.ParseComments,
	}
	if goFiles {
		conf.CreateFromFilenames("adhoc", paths...)
		lprog, err := conf.Load()
		if err != nil {
			log.Fatal(err)
		}
		ps := runner.lint(lprog)
		for _, ps := range ps {
			for _, p := range ps {
				runner.unclean = true
				fmt.Printf("%v: %s\n", p.Position, p.Text)
			}
		}
	} else {
		conf.TypeCheckFuncBodies = func(s string) bool {
			for _, path := range paths {
				if s == path || s == path+"_test" {
					return true
				}
			}
			return false
		}
		for _, path := range paths {
			conf.ImportWithTests(path)
		}
		lprog, err := conf.Load()
		if err != nil {
			log.Fatal(err)
		}
		ps := runner.lint(lprog)
		for _, ps := range ps {
			for _, p := range ps {
				runner.unclean = true
				fmt.Printf("%v: %s\n", p.Position, p.Text)
			}

		}
	}
	if runner.unclean {
		os.Exit(1)
	}
}

func (runner *runner) lint(lprog *loader.Program) map[string][]lint.Problem {
	l := &lint.Linter{
		Checker: runner.checker,
		Ignores: runner.ignores,
	}
	return l.Lint(lprog)
}
