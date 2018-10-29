package server

import (
	"encoding/json"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml"

	yaml "gopkg.in/yaml.v2"
)

type pinnedPkg struct {
	Pkg string
	Rev string
}

// pinnedPkgs is a sorted slice of go pkg names, except always with a `/`
// suffix. This `/` suffix is a sentinel value simplify/reduce the work of
// looking up if we have a pkg.
type pinnedPkgs []pinnedPkg

// Find returns the revision a pkg is pinned at, or the empty string.
func (p pinnedPkgs) Find(pkg string) string {
	if len(p) == 0 {
		return ""
	}
	pkg = pkg + "/"
	i := sort.Search(len(p), func(i int) bool { return p[i].Pkg > pkg })
	if i > 0 && strings.HasPrefix(pkg, p[i-1].Pkg) {
		return p[i-1].Rev
	}
	return ""
}

func (p pinnedPkgs) Len() int           { return len(p) }
func (p pinnedPkgs) Less(i, j int) bool { return p[i].Pkg < p[j].Pkg }
func (p pinnedPkgs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// loadGopkgLock supports unmarshalling the lock file used by
// github.com/golang/dep.
func loadGopkgLock(rawToml []byte) pinnedPkgs {
	lock := struct {
		Projects []struct {
			Name     string `toml:"name"`
			Revision string `toml:"revision"`
			// There are other fields, but we don't use them
		} `toml:"projects"`
		// There are other fields, but we don't use them
	}{}
	err := toml.Unmarshal(rawToml, &lock)
	if err != nil {
		return nil
	}

	pkgs := make(pinnedPkgs, 0, len(lock.Projects))
	for _, p := range lock.Projects {
		pkgs = append(pkgs, pinnedPkg{Pkg: p.Name + "/", Rev: p.Revision})
	}
	sort.Sort(pkgs)
	return pkgs
}

func loadGlideLock(yml []byte) pinnedPkgs {
	type glideLock struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		// There are other fields, but we don't use them
	}
	lock := struct {
		Imports    []*glideLock `yaml:"imports"`
		DevImports []*glideLock `yaml:"testImports"`
		// There are other fields, but we don't use them
	}{}
	err := yaml.Unmarshal(yml, &lock)
	if err != nil {
		return nil
	}

	pkgs := make(pinnedPkgs, 0, len(lock.Imports)+len(lock.DevImports))
	for _, l := range lock.Imports {
		pkgs = append(pkgs, pinnedPkg{Pkg: l.Name + "/", Rev: l.Version})
	}
	for _, l := range lock.DevImports {
		pkgs = append(pkgs, pinnedPkg{Pkg: l.Name + "/", Rev: l.Version})
	}
	sort.Sort(pkgs)
	return pkgs
}

func loadGodeps(b []byte) pinnedPkgs {
	deps := struct {
		Deps []struct {
			ImportPath string
			Rev        string
		}
	}{}
	err := json.Unmarshal(b, &deps)
	if err != nil {
		return nil
	}

	pkgs := make(pinnedPkgs, 0, len(deps.Deps))
	for _, d := range deps.Deps {
		pkgs = append(pkgs, pinnedPkg{Pkg: d.ImportPath + "/", Rev: d.Rev})
	}
	sort.Sort(pkgs)
	return pkgs
}
