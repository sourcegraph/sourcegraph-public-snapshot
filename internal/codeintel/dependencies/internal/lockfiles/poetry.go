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

type Lockfile struct {
	Packages []struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	} `toml:"package"`
}

func ParsePoetryLockFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var lockfile Lockfile
	if _, err := toml.DecodeReader(r, &lockfile); err != nil {
		return nil, errors.Errorf("error decoding poetry lockfile: %w", err)
	}

	var libs []reposource.PackageDependency
	for _, pkg := range lockfile.Packages {
		libs = append(libs, reposource.NewPoetryDependency(pkg.Name, pkg.Version))
	}
	return libs, nil
}
