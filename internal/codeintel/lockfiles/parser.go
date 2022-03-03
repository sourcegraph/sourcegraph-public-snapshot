package lockfiles

import (
	"io"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type parser func(io.Reader) ([]reposource.PackageDependency, error)

var parsers = map[string]parser{
	"package-lock.json": parsePackageLockFile,
	"yarn.lock":         parseYarnLockFile,
}

var lockfilePaths = func() []string {
	paths := make([]string, 0, len(parsers))
	for filename := range parsers {
		paths = append(paths, "*"+filename)
	}
	sort.Strings(paths)

	return paths
}()
