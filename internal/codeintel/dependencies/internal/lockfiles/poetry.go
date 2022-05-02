package lockfiles

import (
	"io"

	"github.com/BurntSushi/toml"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// poetry.lock
//

func parsePoetryLockFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var lockfile struct {
		Packages []struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
		} `toml:"package"`
	}

	if _, err := toml.DecodeReader(r, &lockfile); err != nil {
		return nil, errors.Errorf("error decoding poetry lockfile: %w", err)
	}

	libs := make([]reposource.PackageDependency, 0, len(lockfile.Packages))
	for _, pkg := range lockfile.Packages {
		libs = append(libs, reposource.NewPythonDependency(pkg.Name, pkg.Version))
	}

	return libs, nil
}
