package lockfiles

import (
	"errors"
	"path"
)

type Dependency struct {
	Name    string
	Version string `json:"version"`
	Kind    Kind
}

type Kind string

const (
	KindNPM Kind = "npm"
)

var parsers = map[string]ParseFunc{
	NPMFilename: ParseNPM,
}

type ParseFunc func([]byte) ([]*Dependency, error)

var ErrUnsupported = errors.New("unsupported lockfile kind")

func Parse(file string, data []byte) ([]*Dependency, error) {
	parse, ok := parsers[path.Base(file)]
	if !ok {
		return nil, ErrUnsupported
	}
	return parse(data)
}
