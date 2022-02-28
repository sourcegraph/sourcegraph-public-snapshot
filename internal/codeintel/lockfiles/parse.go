package lockfiles

import (
	"io"
	"path"

	"github.com/aquasecurity/go-dep-parser/pkg/nodejs/npm"
	"github.com/aquasecurity/go-dep-parser/pkg/nodejs/yarn"
	"github.com/aquasecurity/go-dep-parser/pkg/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrUnsupported = errors.New("unsupported lockfile kind")

func Parse(file string, r io.Reader) ([]reposource.PackageDependency, error) {
	p, ok := parsers[path.Base(file)]
	if !ok {
		return nil, ErrUnsupported
	}

	libs, err := p.parse(r)
	if err != nil {
		return nil, err
	}

	var (
		errs errors.MultiError
		deps = make([]reposource.PackageDependency, 0, len(libs))
	)

	for _, lib := range libs {
		dep, err := p.pkg(&lib)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			deps = append(deps, dep)
		}
	}

	return deps, err
}

var parsers = map[string]*parser{
	"package-lock.json": {npm.Parse, npmPackage},
	"yarn.lock":         {yarn.Parse, npmPackage},
}

type parser struct {
	parse func(io.Reader) ([]types.Library, error)
	pkg   func(*types.Library) (reposource.PackageDependency, error)
}

func npmPackage(lib *types.Library) (reposource.PackageDependency, error) {
	return reposource.ParseNPMDependency(lib.Name + "@" + lib.Version)
}
