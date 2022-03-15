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

// lockfilePathspecs is the list of git pathspecs that match lockfiles.
//
// https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-aiddefpathspecapathspec
var lockfilePathspecs = func() []string {
	paths := make([]string, 0, len(parsers))
	for filename := range parsers {
		paths = append(paths, "*"+filename)
	}
	sort.Strings(paths)

	return paths
}()
