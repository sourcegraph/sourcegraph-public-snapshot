package lockfiles

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const NPMFilename = "package-lock.json"

func ParseNPM(b []byte) (deps []reposource.PackageDependency, err error) {
	var lockfile struct {
		Dependencies map[string]*struct{ Version string }
	}

	err = json.Unmarshal(b, &lockfile)
	if err != nil {
		return nil, err
	}

	var errs errors.MultiError

	// TODO: Make json decoder unmarshal dependencies in order.
	for name, d := range lockfile.Dependencies {
		dep, err := reposource.ParseNPMDependency(name + "@" + d.Version)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			deps = append(deps, dep)
		}
	}

	return deps, err
}
