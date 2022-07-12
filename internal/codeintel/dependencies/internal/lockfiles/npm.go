package lockfiles

import (
	"encoding/json"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// package-lock.json
//

type packageLockDependency struct {
	Version      string
	Dev          bool
	Dependencies map[string]*packageLockDependency
}

func parsePackageLockFile(r io.Reader) ([]reposource.PackageVersion, error) {
	var lockFile struct {
		Dependencies map[string]*packageLockDependency
	}

	err := json.NewDecoder(r).Decode(&lockFile)
	if err != nil {
		return nil, errors.Errorf("decode error: %w", err)
	}

	return parsePackageLockDependencies(lockFile.Dependencies)
}

func parsePackageLockDependencies(in map[string]*packageLockDependency) ([]reposource.PackageVersion, error) {
	var (
		errs errors.MultiError
		out  = make([]reposource.PackageVersion, 0, len(in))
	)

	for name, d := range in {
		dep, err := reposource.ParseNpmPackageVersion(name + "@" + d.Version)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			out = append(out, dep)
		}

		if d.Dependencies != nil {
			// Recursion
			deps, err := parsePackageLockDependencies(d.Dependencies)
			out = append(out, deps...)
			errs = errors.Append(errs, err)
		}
	}

	return out, errs
}
