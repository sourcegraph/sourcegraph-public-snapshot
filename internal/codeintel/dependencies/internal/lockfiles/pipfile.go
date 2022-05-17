package lockfiles

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// Pipfile.lock
//

// parsePipfileLockFile extracts all dependencies, both "default" and "develop", from a Pipfile.lock file.
//
// The Pipfile project (https://github.com/pypa/pipfile) looks abandoned and redirects to another
// repo which comes, luckily, from the same organization: https://github.com/pypa/pipenv
// It also claims that this pipenv repo contains the reference implementation for the Pipfile.lock
// generator.
//
// From there we infer that:
// 1. Pipfile.lock consists of a JSON object with 3 elements, "_meta", "default", and "develop",
// each of which has zero or more members inside, so basically a map of maps.
// 2. The "default" and "develop" objects list all the dependencies of the project,
// including transitive ones.
//
// Further discussion at https://github.com/sourcegraph/sourcegraph/issues/35041
func parsePipfileLockFile(r io.Reader) ([]reposource.PackageDependency, error) {
	type packageInfo struct {
		Version string
	}

	var lockfile struct {
		Default map[string]packageInfo
		Develop map[string]packageInfo
	}

	err := json.NewDecoder(r).Decode(&lockfile)
	if err != nil {
		return nil, errors.Errorf("error decoding Pipfile.lock: %w", err)
	}

	libs := make([]reposource.PackageDependency, 0, len(lockfile.Default)+len(lockfile.Develop))
	for pkgName, info := range lockfile.Default {
		libs = append(libs, reposource.NewPythonDependency(pkgName, strings.TrimPrefix(info.Version, "==")))
	}
	for pkgName, info := range lockfile.Develop {
		libs = append(libs, reposource.NewPythonDependency(pkgName, strings.TrimPrefix(info.Version, "==")))
	}
	return libs, nil
}
