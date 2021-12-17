package catalog

import (
	"context"
	"encoding/json"
	"os"

	"golang.org/x/tools/go/packages"
)

func GetPackages(ctx context.Context) (pkgs []Package, _ error) {
	goMods, err := getAllGoModules(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range goMods {
		pkgs = append(pkgs, Package{Name: m.Path})
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

type goModuleInfo struct {
	*packages.Module
	Packages []*packages.Package
}

func getAllGoModules(ctx context.Context) ([]*goModuleInfo, error) {
	const cacheFile = "/tmp/sqs-wip-cache/all-goModuleInfo.json"
	_ = os.MkdirAll(filepath.Dir(cacheFile), 0700)
	f, err := os.Open(cacheFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		defer f.Close()
		var data []*goModuleInfo
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return nil, err
		}
		return data, nil
	}

	const dir = "/home/sqs/src/github.com/sourcegraph/sourcegraph.tmp" // TODO(sqs)

	cfg := &packages.Config{
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedModule,
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

	modulesByPath := map[string]*goModuleInfo{}
	for _, pkg := range allPkgs {
		if pkg.Module == nil {
			continue
		}
		// TODO(sqs): handle (*packages.Module).Replace, etc.
		key := pkg.Module.Path
		info := modulesByPath[key]
		if info == nil {
			info = &goModuleInfo{Module: pkg.Module}
			modulesByPath[key] = info
		}
		info.Packages = append(info.Packages, pkg)
	}
	allModules := make([]*goModuleInfo, 0, len(modulesByPath))
	for _, mod := range modulesByPath {
		allModules = append(allModules, mod)
	}

	f, err = os.Create(cacheFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(allModules); err != nil {
		return nil, err
	}
	return allModules, nil
}
