// Package plan configures the CI test plan for a repository.
package plan

import (
	"net/url"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

// Create creates a CI test plan, given a .drone.yml file. It returns
// the final .drone.yml to use, after adding inferred build steps,
// srclib analysis, srclib import, etc.
//
// Create should execute very quickly and must not make any server API
// calls, since it can be run locally.
func Create(configYAML string, droneYMLFileExists bool, inv *inventory.Inventory, srclibImportURL, srclibCoverageURL *url.URL) (string, []matrix.Axis, error) {
	config, err := droneyaml.Parse([]byte(configYAML))
	if err != nil {
		return "", nil, err
	}
	axes, err := parseMatrix(configYAML)
	if err != nil {
		return "", nil, err
	}

	if !droneYMLFileExists {
		// Generate a reasonable default configuration.
		var err error
		config, axes, err = inferConfig(inv)
		if err != nil {
			return "", nil, err
		}
	}

	// Add the srclib analysis steps to the CI test plan.
	if err := configureSrclib(inv, config, axes, srclibImportURL, srclibCoverageURL); err != nil {
		return "", nil, err
	}

	if err := setCloneCompleteHistory(config); err != nil {
		return "", nil, err
	}

	finalConfigYAML, err := marshalConfigWithMatrix(*config, axes)
	if err != nil {
		return "", nil, err
	}
	return string(finalConfigYAML), axes, nil
}

// setCloneCompleteHistory sets the clone step to NOT use a shallow clone so
// that all of the refspec's commits get fetched. This is necessary
// for us to build the 2000th-old commit on a branch, for example (git
// fetch --depth 50, which is Drone CI's drone-git's default, would
// always omit it).
func setCloneCompleteHistory(config *droneyaml.Config) error {
	if config.Clone.Vargs == nil {
		config.Clone.Vargs = droneyaml.Vargs{}
	}

	config.Clone.Vargs["complete"] = true
	return nil
}
