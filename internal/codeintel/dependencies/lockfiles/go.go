package lockfiles

import (
	"io"

	"github.com/inconshreveable/log15"
	"golang.org/x/mod/modfile"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// go.mod
//

// based on https://go.dev/ref/mod#go-mod-file.
func parseGoModFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var (
		errs    errors.MultiError
		deps    []reposource.PackageDependency
		ignore  = make(map[string]string)
		replace = make(map[string]*modfile.Replace)
	)

	data := make([]byte, 1024*1024)
	n, err := r.Read(data)
	if err == io.EOF {
		if n == 0 {
			return nil, nil
		}
		err = nil
	}
	if err != nil {
		return nil, err
	}

	// We log the size of go.mod files to get an understanding of their distribution.
	// We can remove this log later.
	log15.Debug("gomod", "size", n)

	data = data[0:n]
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

		path := r.Mod.Path
		version := r.Mod.Version

		if v, ok := replace[path]; ok && (v.Old.Version == "" || v.Old.Version == r.Mod.Version) {
			path = v.New.Path
			version = v.New.Version
		}

		dep, err := reposource.ParseGoDependency(path + "@" + version)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			deps = append(deps, dep)
		}
	}
	return deps, errs
}
