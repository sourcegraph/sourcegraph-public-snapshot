package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

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

	npmMods, err := getAllNpmModules(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range npmMods {
		pkgs = append(pkgs, Package{Name: m.Name})
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

type npmModuleInfo struct {
	Name string
}

func getAllNpmModules(ctx context.Context) ([]*npmModuleInfo, error) {
	const cacheFile = "/tmp/sqs-wip-cache/all-npmModuleInfo.json"
	_ = os.MkdirAll(filepath.Dir(cacheFile), 0700)

	f, err := os.Open(cacheFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		defer f.Close()
		var data []*npmModuleInfo
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return nil, err
		}
		return data, nil
	}

	var yarnLockfiles []string
	const dir = "/home/sqs/src/github.com/sourcegraph/sourcegraph.tmp" // TODO(sqs)
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if filepath.Base(path) == "node_modules" {
			return filepath.SkipDir
		}
		if !d.IsDir() && filepath.Base(path) == "yarn.lock" {
			yarnLockfiles = append(yarnLockfiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var allModules []*npmModuleInfo
	seen := map[string]struct{}{}
	for _, yarnLockfile := range yarnLockfiles {
		data, err := os.ReadFile(yarnLockfile)
		if err != nil {
			return nil, err
		}
		entries, err := parseYarnLockfile(data)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if _, seen := seen[e.Name]; seen {
				continue
			}
			seen[e.Name] = struct{}{}
			allModules = append(allModules, &npmModuleInfo{Name: e.Name})
		}
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

type YarnLockfileEntry struct {
	Name             string
	Version          string
	NameVersionSpecs []string
	// TODO(sqs)
}

func parseYarnLockfile(data []byte) (entries []*YarnLockfileEntry, err error) {
	lines := bytes.Split(data, []byte("\n"))
	var cur *YarnLockfileEntry
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		if line[0] != ' ' {
			cur = &YarnLockfileEntry{}
			entries = append(entries, cur)

			// TODO(sqs): hacky
			lineWithoutTrailingColon := line[:len(line)-1]
			nameVersionSpecs := bytes.Split(lineWithoutTrailingColon, []byte(","))
			for _, spec := range nameVersionSpecs {
				spec = bytes.TrimPrefix(bytes.TrimSuffix(bytes.TrimSpace(spec), []byte{'"'}), []byte{'"'})
				cur.NameVersionSpecs = append(cur.NameVersionSpecs, string(spec))
			}

			// TODO(sqs): assume that first name version spec's name is the entry name? is this
			// always true?
			cur.Name = cur.NameVersionSpecs[0][:strings.LastIndex(cur.NameVersionSpecs[0], "@")]
		} else if versionPrefix := "  version \""; bytes.HasPrefix(line, []byte(versionPrefix)) {
			version := line[len(versionPrefix) : len(line)-1]
			cur.Version = string(version)
		}
	}

	return entries, nil
}
