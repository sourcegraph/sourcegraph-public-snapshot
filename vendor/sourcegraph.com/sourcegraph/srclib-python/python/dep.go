package python

import (
	"encoding/json"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/dep"
)

func ResolveDep(d interface{}) (*dep.ResolvedTarget, error) {
	req, err := asRequirement(d)
	if err != nil {
		return nil, err
	}

	specStrings := make([]string, len(req.Specs))
	for i, spec := range req.Specs {
		specStrings[i] = spec[0] + " " + spec[1]
	}

	return &dep.ResolvedTarget{
		ToRepoCloneURL:  req.RepoURL,
		ToUnit:          req.ProjectName,
		ToUnitType:      DistPackageSourceUnitType,
		ToVersionString: strings.Join(specStrings, ", "),
		ToRevSpec:       "",
	}, nil
}

// Kludge helper function, because dep gets deserialized as a map[string]interface{} rather than a requirement
func asRequirement(dep interface{}) (*requirement, error) {
	var req requirement

	b, err := json.Marshal(dep)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}
