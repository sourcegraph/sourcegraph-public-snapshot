package lockfiles

import (
	"io"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type parser func(io.Reader) ([]reposource.PackageDependency, error)

var parsers = map[string]parser{
	"package-lock.json": parsePackageLockFile,
	"yarn.lock":         parseYarnLockFile,
	"go.mod":            parseGoModFile,
	"poetry.lock":       parsePoetryLockFile,
	"Pipfile.lock":      parsePipfileLockFile,
}

// lockfilePathspecs is the list of git pathspecs that match lockfiles.
//
// https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-aiddefpathspecapathspec
var lockfilePathspecs = func() []gitserver.Pathspec {
	pathspecs := make([]gitserver.Pathspec, 0, len(parsers))
	for filename := range parsers {
		pathspecs = append(pathspecs, gitserver.PathspecSuffix(filename))
	}
	return pathspecs
}()
