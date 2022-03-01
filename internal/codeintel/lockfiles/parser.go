package lockfiles

import (
	"io"
	"sort"

	"github.com/aquasecurity/go-dep-parser/pkg/nodejs/npm"
	"github.com/aquasecurity/go-dep-parser/pkg/nodejs/yarn"
	"github.com/aquasecurity/go-dep-parser/pkg/types"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type parser struct {
	parseLockfileContents               func(io.Reader) ([]types.Library, error)
	transformLibraryToPackageDependency func(types.Library) (reposource.PackageDependency, error)
}

var parsers = map[string]*parser{
	"package-lock.json": {npm.Parse, npmPackage},
	"yarn.lock":         {yarn.Parse, npmPackage},
}

var lockfilePaths = func() []string {
	paths := make([]string, 0, len(parsers))
	for filename := range parsers {
		paths = append(paths, "*"+filename)
	}
	sort.Strings(paths)

	return paths
}()

func npmPackage(lib types.Library) (reposource.PackageDependency, error) {
	return reposource.ParseNPMDependency(lib.Name + "@" + lib.Version)
}
