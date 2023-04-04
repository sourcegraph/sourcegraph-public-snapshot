package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/run"

	depgraph "github.com/sourcegraph/sourcegraph/dev/depgraph/internal/graph"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	summaryFlagSet  = flag.NewFlagSet("depgraph summary", flag.ExitOnError)
	summaryDepsSum  = summaryFlagSet.Bool("deps.sum", false, "generate md5sum of each dependency")
	summaryDepsOnly = summaryFlagSet.Bool("deps.only", false, "only display dependencies")
)

var summaryCommand = &ffcli.Command{
	Name:       "summary",
	ShortUsage: "depgraph summary {package}",
	ShortHelp:  "Outputs a text summary of the given package dependency and dependents",
	FlagSet:    summaryFlagSet,
	Exec:       summary,
}

func summary(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.Errorf("expected exactly one package")
	}
	pkg := args[0]

	root, err := findRoot()
	if err != nil {
		return err
	}

	graph, err := depgraph.Load(root)
	if err != nil {
		return err
	}
	if _, ok := graph.PackageNames[pkg]; !ok {
		return errors.Newf("pkg %q not found", pkg)
	}

	dependencyMap := summaryTraverse(pkg, graph.Dependencies)
	dependencies := make([]string, 0, len(dependencyMap))
	for dependency := range dependencyMap {
		dependencies = append(dependencies, dependency)
	}
	sort.Strings(dependencies)

	dependentMap := summaryTraverse(pkg, graph.Dependents)
	dependents := make([]string, 0, len(dependentMap))
	for dependent := range dependentMap {
		dependents = append(dependents, dependent)
	}
	sort.Strings(dependents)

	fmt.Printf("Target package:\n")
	printPkg(ctx, root, pkg)

	fmt.Printf("\n")
	fmt.Printf("Direct dependencies:\n")

	for _, dependency := range dependencies {
		if dependencyMap[dependency] {
			printPkg(ctx, root, dependency)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Transitive dependencies:\n")

	for _, dependency := range dependencies {
		if !dependencyMap[dependency] {
			printPkg(ctx, root, dependency)
		}
	}

	if *summaryDepsOnly {
		return nil
	}

	fmt.Printf("\n")
	fmt.Printf("Dependent commands:\n")

	for _, dependent := range dependents {
		if isMain(graph, dependent) {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Direct dependents:\n")

	for _, dependent := range dependents {
		if !isMain(graph, dependent) && dependentMap[dependent] {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Transitive dependents:\n")

	for _, dependent := range dependents {
		if !isMain(graph, dependent) && !dependentMap[dependent] {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	return nil
}

// summaryTraverse returns a set of packages related to the given package via the given
// relation. Each package is returned with a boolean value indicating whether or not the
// relation is direct (true) or transitive (false).k
func summaryTraverse(pkg string, relation map[string][]string) map[string]bool {
	m := make(map[string]bool, len(relation[pkg]))
	for _, v := range relation[pkg] {
		m[v] = true
	}

outer:
	for {
		for k := range m {
			for _, v := range relation[k] {
				if _, ok := m[v]; ok {
					continue
				}

				m[v] = false
				continue outer
			}
		}

		break
	}

	return m
}

// isMain returns true if the given package declares "main" in the given package name map.
func isMain(graph *depgraph.DependencyGraph, pkg string) bool {
	for _, name := range graph.PackageNames[pkg] {
		if name == "main" {
			return true
		}
	}

	return false
}

func printPkg(ctx context.Context, root string, pkg string) error {
	fmt.Printf("\t> %s", pkg)
	if *summaryDepsSum {
		dir := "./" + pkg
		lines, err := run.Bash(ctx, "tar c", dir, "| md5sum").
			Dir(root).
			Run().
			Lines()
		if err != nil {
			return err
		}
		sum := strings.Split(lines[0], " ")[0]
		fmt.Printf("\t%s", sum)
	}
	fmt.Println()
	return nil
}
