package lockfiles

import (
	"fmt"
	"path"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Dependency struct {
	Name    string
	Version string
	Kind    Kind
}

func (d *Dependency) Less(other *Dependency) bool {
	if d.Kind != other.Kind {
		return d.Kind < other.Kind
	}

	if d.Name != other.Name {
		return d.Name < other.Name
	}

	return d.Version < other.Version
}

func (d *Dependency) String() string {
	return fmt.Sprintf("%s %s %s", d.Kind, d.Name, d.Version)
}

type Kind string

const (
	KindNPM Kind = "npm"
)

var parsers = map[string]ParseFunc{
	NPMFilename: ParseNPM,
}

type ParseFunc func([]byte) ([]reposource.PackageDependency, error)

var ErrUnsupported = errors.New("unsupported lockfile kind")

func Parse(file string, data []byte) ([]reposource.PackageDependency, error) {
	parse, ok := parsers[path.Base(file)]
	if !ok {
		return nil, ErrUnsupported
	}
	return parse(data)
}
