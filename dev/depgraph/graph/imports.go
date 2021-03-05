package graph

import (
	"sort"
	"strings"
	"sync"
)

var concurrencyLevel = 1

// importsOfPackages runs importsOfPackage on each of the given packages concurrently
// and returns a map from packages to the set of internal packages it improts.
func importsOfPackages(pkgs []string) (map[string][]string, error) {
	ch := make(chan string, len(pkgs))
	for _, pkg := range pkgs {
		ch <- pkg
	}
	close(ch)

	type pair struct {
		pkg     string
		imports []string
		err     error
	}

	var wg sync.WaitGroup
	pairs := make(chan pair, len(pkgs))

	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for pkg := range ch {
				imports, err := importsOfPackage(pkg)
				pairs <- pair{pkg, imports, err}
			}
		}()
	}
	wg.Wait()
	close(pairs)

	allImports := make(map[string][]string, len(pkgs))
	for pair := range pairs {
		if err := pair.err; err != nil {
			return nil, err
		}

		allImports[pair.pkg] = pair.imports
	}

	return allImports, nil
}

// importsOfPackage returns an ordered list of packages imported by the given package.
// This includes only packages that are defined within the sourcegraph/sourcegraph repo.
func importsOfPackage(pkg string) ([]string, error) {
	importTemplates := []string{
		`{{ join .Imports "\n" }}`,
		`{{ join .TestImports "\n" }}`,
	}

	pkgMap := map[string]struct{}{}
	for _, template := range importTemplates {
		out, err := runGo("list", "-f", template, RootPackage+"/"+pkg)
		if err != nil {
			return nil, err
		}

		for _, pkg := range strings.Split(out, "\n") {
			if strings.HasPrefix(pkg, RootPackage) {
				pkgMap[pkg] = struct{}{}
			}
		}
	}

	packages := make([]string, 0, len(pkgMap))
	for pkg := range pkgMap {
		packages = append(packages, trimPackage(pkg))
	}
	sort.Strings(packages)

	return packages, nil
}
