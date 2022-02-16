package lockfiles

import (
	"encoding/json"
	"sort"
)

type Dependency struct {
	Name    string
	Version string `json:"version"`
}

func ParseNpm(b []byte) ([]*Dependency, error) {
	var lockfile struct {
		Dependencies map[string]*Dependency `json:"dependencies"`
	}

	err := json.Unmarshal(b, &lockfile)
	if err != nil {
		return nil, err
	}

	var dependencies []*Dependency
	for name, dependency := range lockfile.Dependencies {
		dependency.Name = name
		dependencies = append(dependencies, dependency)
	}

	sort.SliceStable(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})

	return dependencies, nil
}
