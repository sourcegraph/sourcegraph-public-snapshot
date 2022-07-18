package lockfiles

import (
	"io"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type parser func(io.Reader) ([]reposource.VersionedPackage, *DependencyGraph, error)

type nonGraphParser func(io.Reader) ([]reposource.VersionedPackage, error)

func wrapNonGraphParser(f nonGraphParser) parser {
	return func(r io.Reader) ([]reposource.VersionedPackage, *DependencyGraph, error) {
		deps, err := f(r)
		return deps, nil, err
	}
}

var parsers = map[string]parser{
	"package-lock.json": wrapNonGraphParser(parsePackageLockFile),
	"yarn.lock":         parseYarnLockFile,
	"go.mod":            wrapNonGraphParser(parseGoModFile),
	"poetry.lock":       wrapNonGraphParser(parsePoetryLockFile),
	"Pipfile.lock":      wrapNonGraphParser(parsePipfileLockFile),
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
