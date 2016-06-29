package plan

import (
	"fmt"
	"strings"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

// inferConfig consults a repo's inventory (of programming languages
// used) and generates a .sg-drone.yml file that will build the
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
			// Only warn about languages which are possible to build.
			if lang.Type == "programming" {
				unsupported = append(unsupported, lang.Name)
			}
		} else {
			config.Build = append(config.Build, c.build)
		}
		for key, vals := range c.matrix {
			matrix[key] = append(matrix[key], vals...)
		}
	}

	if len(config.Build) == 0 {
		config.Build = append(config.Build, buildLogMsg("Couldn't infer CI build config; please create a .sg-drone.yml file", "no supported programming languages were auto-detected"))
	}

	if len(unsupported) > 0 {
		config.Build = append(config.Build, buildLogMsg(
			fmt.Sprintf("Can't automatically generate CI build config for %s; please create a .sg-drone.yml file", strings.Join(unsupported, ", ")),
			fmt.Sprintf("automatic CI config does not yet support:\n%s\n", strings.Join(unsupported, "\n")),
		))
	}

	return &config, calcMatrix(matrix), nil
}

var langConfigs = map[string]struct {
	build  droneyaml.BuildItem
	matrix map[string][]string
}{
	"JavaScript": {},
	"Go":         {},
	"Java":       {},
	"Python":     {},
	"TypeScript": {},
	"C#":         {},
	"CSS":        {},
}
