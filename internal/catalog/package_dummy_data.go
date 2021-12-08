package catalog

import (
	"context"
	"encoding/json"
	"os"

	"golang.org/x/tools/go/packages"
)

func GetPackages(ctx context.Context) ([]Package, error) {
	allPkgs, err := getAllPackages(ctx)
	if err != nil {
		return nil, err
	}

	pkgs := make([]Package, len(allPkgs))
	for i, pkg := range allPkgs {
		pkgs[i] = Package{Name: pkg.ID}
	}
	return pkgs, nil
}

func AllPackages() []Package {
	pkgs, err := GetPackages(nil)
	if err != nil {
		panic(err)
	}
	return pkgs
}

func getAllPackages(ctx context.Context) ([]*packages.Package, error) {
	const cacheFile = "/tmp/sqs-wip-cache/allPackages.json"
	f, err := os.Open(cacheFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		defer f.Close()
		var data []*packages.Package
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return nil, err
		}
		return data, nil
	}

	const dir = "/home/sqs/src/github.com/sourcegraph/sourcegraph.tmp" // TODO(sqs)

	cfg := &packages.Config{
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedImports,
		Context: ctx,
		Dir:     dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	allPkgs := make([]*packages.Package, 0, len(pkgs))
	packages.Visit(pkgs, func(p *packages.Package) bool {
		allPkgs = append(allPkgs, p)
		return true
	}, nil)

	f, err = os.Create(cacheFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(allPkgs); err != nil {
		return nil, err
	}
	return allPkgs, nil
}
