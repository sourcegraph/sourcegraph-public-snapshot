package lockfiles

import (
	"io"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

//
// go.mod
//

// based on https://go.dev/ref/mod#go-mod-file.
func parseGoModFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var (
		deps    []reposource.PackageDependency
		ignore  = make(map[string]string)
		replace = make(map[string]*modfile.Replace)
	)

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range f.Replace {
		// According to https://go.dev/ref/mod#go-mod-file-replace, the replacement
		// version is empty if and only if the replacement path is local.
		if r.New.Version == "" {
			// Ignore dependencies which point to local modules.
			ignore[r.Old.Path] = r.Old.Version
		} else {
			replace[r.Old.Path] = r
		}
	}

	for _, r := range f.Exclude {
		ignore[r.Mod.Path] = r.Mod.Version
	}

	for _, r := range f.Require {
		if v, ok := ignore[r.Mod.Path]; ok && v == r.Mod.Version {
			continue
		}

		v := module.Version{
			Path:    r.Mod.Path,
			Version: r.Mod.Version,
		}

		if s, ok := replace[r.Mod.Path]; ok && (s.Old.Version == "" || s.Old.Version == r.Mod.Version) {
			v.Path = s.New.Path
			v.Version = s.New.Version
		}

		deps = append(deps, reposource.NewGoDependency(v))
	}
	return deps, nil
}
