package plan

import (
	"fmt"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
)

// autogenerateConfig consults a repo's inventory (of programming
// languages used) and generates a .drone.yml file that will build the
// repo. This is not guaranteed to be correct. The primary purpose of
// this auto-generated config is to fetch the project dependencies and
// compile the project to prepare for the srclib analysis step.
func autogenerateConfig(inv *inventory.Inventory) (*droneyaml.Config, []matrix.Axis, error) {
	// Merge the default configs for all of the languages we detect.
	var config droneyaml.Config
	matrix := matrix.Matrix{}
	for _, lang := range inv.Languages {
		c, ok := langConfigs[lang.Name]
		if !ok {
			c.builds = []droneyaml.BuildItem{
				buildLogMsg(fmt.Sprintf("Can't automatically generate CI build config for %s; please create a .drone.yml file", lang.Name), fmt.Sprintf("automatic CI config does not yet support %s", lang.Name)),
			}
		}

		config.Build = append(config.Build, c.builds...)
		for key, vals := range c.matrix {
			matrix[key] = append(matrix[key], vals...)
		}
	}

	if len(config.Build) == 0 {
		config.Build = append(config.Build, buildLogMsg("Couldn't infer CI build config; please create a .drone.yml file", "no supported programming languages were auto-detected"))
	}

	return &config, calcMatrix(matrix), nil
}

var langConfigs = map[string]struct {
	builds []droneyaml.BuildItem
	matrix map[string][]string
}{
	"Go": {
		builds: []droneyaml.BuildItem{
			{
				Key: "Go $$GO_VERSION build",
				Build: droneyaml.Build{
					Container: droneyaml.Container{Image: "golang:$$GO_VERSION"},
					Commands: []string{
						"go get -t ./...",
						"go build ./...",
					},
				},
			},
			{
				Key: "Go $$GO_VERSION test",
				Build: droneyaml.Build{
					Container: droneyaml.Container{Image: "golang:$$GO_VERSION"},
					Commands: []string{
						"go test -v ./...",
					},
					AllowFailure: true,
				},
			},
		},
		matrix: map[string][]string{"GO_VERSION": []string{"1.5"}},
	},
	"JavaScript": {
		builds: []droneyaml.BuildItem{
			{
				Key: "JavaScript deps (node v$$NODE_VERSION)",
				Build: droneyaml.Build{
					Container: droneyaml.Container{Image: "node:$$NODE_VERSION"},
					Commands: []string{
						"npm install --quiet",
					},
				},
			},
			{
				Key: "JavaScript test (node v$$NODE_VERSION)",
				Build: droneyaml.Build{
					Container: droneyaml.Container{Image: "node:$$NODE_VERSION"},
					Commands: []string{
						"npm run test",
					},
					AllowFailure: true,
				},
			},
		},
		matrix: map[string][]string{"NODE_VERSION": []string{"4"}},
	},
}
