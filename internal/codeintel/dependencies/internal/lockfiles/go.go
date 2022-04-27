package lockfiles

import (
	"io"

	"golang.org/x/mod/modfile"

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
			// Ignore dependencies which point to local modules. r.Old.Version might be ""
			// which means that all version will be replaced.
			ignore[r.Old.Path] = r.Old.Version
		} else {
			replace[r.Old.Path] = r
		}
	}

	for _, r := range f.Exclude {
		ignore[r.Mod.Path] = r.Mod.Version
	}

	for _, r := range f.Require {
		if s, ok := ignore[r.Mod.Path]; ok && (s == "" || s == r.Mod.Version) {
			continue
		}

		if s, ok := replace[r.Mod.Path]; ok && (s.Old.Version == "" || s.Old.Version == r.Mod.Version) {
			r.Mod.Path = s.New.Path
			r.Mod.Version = s.New.Version
		}

		deps = append(deps, reposource.NewGoDependency(r.Mod))
	}
	return deps, nil
}
