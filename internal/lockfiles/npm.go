package lockfiles

import (
	"encoding/json"
	"sort"
)

const NPMFilename = "package-lock.json"

func ParseNPM(b []byte) ([]*Dependency, error) {
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
		dependency.Kind = KindNPM
		dependencies = append(dependencies, dependency)
	}

	// TODO: We want to use the json decoder to unmarshal dependencies in
	// order rather than having to sort here.
	sort.SliceStable(dependencies, func(i, j int) bool {
		return dependencies[i].Less(dependencies[j])
	})

	return dependencies, nil
}
