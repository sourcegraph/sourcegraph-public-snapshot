package plan

import (
	"fmt"
	"strings"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

// inferConfig consults a repo's inventory (of programming languages
// used) and generates a .drone.yml file that will build the
// repo. This is not guaranteed to be correct. The primary purpose of
// this inferred config is to fetch the project dependencies and
// compile the project to prepare for the srclib analysis step.
func inferConfig(inv *inventory.Inventory) (*droneyaml.Config, []matrix.Axis, error) {
	// Merge the default configs for all of the languages we detect.
	var config droneyaml.Config
	unsupported := []string{}
	matrix := matrix.Matrix{}
	for _, lang := range inv.Languages {
		c, ok := langConfigs[lang.Name]
		if !ok {
			unsupported = append(unsupported, lang.Name)
		} else {
			config.Build = append(config.Build, c.build)
		}
		for key, vals := range c.matrix {
			matrix[key] = append(matrix[key], vals...)
		}
	}

	if len(config.Build) == 0 {
		config.Build = append(config.Build, buildLogMsg("Couldn't infer CI build config; please create a .drone.yml file", "no supported programming languages were auto-detected"))
	}

	if len(unsupported) > 0 {
		config.Build = append(config.Build, buildLogMsg(
			fmt.Sprintf("Can't automatically generate CI build config for %s; please create a .drone.yml file", strings.Join(unsupported, ", ")),
			fmt.Sprintf("automatic CI config does not yet support:\n%s\n", strings.Join(unsupported, "\n")),
		))
	}

	return &config, calcMatrix(matrix), nil
}

var langConfigs = map[string]struct {
	build  droneyaml.BuildItem
	matrix map[string][]string
}{
	"Go": {},
	"JavaScript": {
		build: droneyaml.BuildItem{
			Key: "JavaScript deps (node v$$NODE_VERSION)",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "node:$$NODE_VERSION"},
				Commands: []string{
					// If the required package.json file is not in the root
					// directory, attempt to find and navigate to it within
					// subdirectories, excluding any node_modules directories.
					`[ -f package.json ] || cd "$(dirname "$(find ./ -type f -name package.json -not -path '*/node_modules/*' | tail -1)")"`,
					"[ -f package.json ] && npm install --quiet",
				},
				AllowFailure: true,
			},
		},
		matrix: map[string][]string{"NODE_VERSION": {"4"}},
	},
	"Java":       {},
	"Python":     {},
	"TypeScript": {},
	"C#":         {},
	"CSS":        {},
}
