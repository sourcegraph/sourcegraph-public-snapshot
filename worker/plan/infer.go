package plan

import (
	"fmt"
	"os"
	"strconv"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
)

var dontInferTestSteps, _ = strconv.ParseBool(os.Getenv("SRC_DONT_INFER_TEST_STEPS"))

// inferConfig consults a repo's inventory (of programming languages
// used) and generates a .drone.yml file that will build the
// repo. This is not guaranteed to be correct. The primary purpose of
// this inferred config is to fetch the project dependencies and
// compile the project to prepare for the srclib analysis step.
func inferConfig(inv *inventory.Inventory) (*droneyaml.Config, []matrix.Axis, error) {
	// Merge the default configs for all of the languages we detect.
	var config droneyaml.Config
	matrix := matrix.Matrix{}
	for _, lang := range inv.Languages {
		c, ok := langConfigs[lang.Name]
		if !ok {
			c.build = buildLogMsg(fmt.Sprintf("Can't automatically generate CI build config for %s; please create a .drone.yml file", lang.Name), fmt.Sprintf("automatic CI config does not yet support %s", lang.Name))
		}

		config.Build = append(config.Build, c.build)
		if !dontInferTestSteps {
			config.Build = append(config.Build, c.test)
		}
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
	build  droneyaml.BuildItem
	test   droneyaml.BuildItem
	matrix map[string][]string
}{
	"Go": {
		build: droneyaml.BuildItem{
			Key: "Go $$GO_VERSION build",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "golang:$$GO_VERSION"},
				Commands: []string{
					"go get -t ./...",
					"go build ./...",
				},
				AllowFailure: true,
			},
		},
		test: droneyaml.BuildItem{
			Key: "Go $$GO_VERSION test",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "golang:$$GO_VERSION"},
				Commands: []string{
					"go test -v ./...",
				},
				AllowFailure: true,
			},
		},
		matrix: map[string][]string{"GO_VERSION": []string{"1.5"}},
	},
	"JavaScript": {
		build: droneyaml.BuildItem{
			Key: "JavaScript deps (node v$$NODE_VERSION)",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "node:$$NODE_VERSION"},
				Commands: []string{
					// If the required package.json file is not in the root directory,
					// attempt to find and navigate to it within subdirectories.
					`[ -f package.json ] || cd "$(dirname "$(find ./ -type f -name package.json | head -1)")"`,
					"[ -f package.json ] && npm install --quiet",
				},
				AllowFailure: true,
			},
		},
		test: droneyaml.BuildItem{
			Key: "JavaScript test (node v$$NODE_VERSION)",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "node:$$NODE_VERSION"},
				Commands: []string{
					`[ -f package.json ] || cd "$(dirname "$(find ./ -type f -name package.json | head -1)")"`,
					"[ -f package.json ] && npm run test",
				},
				AllowFailure: true,
			},
		},
		matrix: map[string][]string{"NODE_VERSION": []string{"4"}},
	},
	"Java": {
		build: droneyaml.BuildItem{
			Key: "Java build (Java $$JAVA_VERSION)",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "maven:3-jdk-$$JAVA_VERSION"},
				Commands: []string{
					"[ -f pom.xml ] && mvn --quiet package",
					"[ -f build.gradle ] && (([ -f gradlew ] && ./gradlew build) || gradle build)",
				},
				AllowFailure: true,
			},
		},
		test: droneyaml.BuildItem{
			Key: "Java test (Java $$JAVA_VERSION)",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "maven:3-jdk-$$JAVA_VERSION"},
				Commands: []string{
					"[ -f pom.xml ] && mvn --quiet test",
					"[ -f build.gradle ] && (([ -f gradlew ] && ./gradlew test) || gradle test)",
				},
				AllowFailure: true,
			},
		},
		matrix: map[string][]string{"JAVA_VERSION": []string{"8"}},
	},
}
