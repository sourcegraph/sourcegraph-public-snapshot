package plan

import (
	"fmt"
	"net/url"
	"strings"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

// configureSrclib modifies the Drone config to run srclib analysis
// during the CI build.
func configureSrclib(inv *inventory.Inventory, config *droneyaml.Config, axes []matrix.Axis, srclibImportURL *url.URL) error {
	var srclibExplicitlyConfigured bool
	for _, step := range config.Build {
		// Rough heuristic for now: does the Docker image name contain
		// "srclib".
		// (alexsaveliev) excluding srclib-java from heuristic
		// because it's used both by build/test and srclib steps
		if strings.Contains(step.Container.Image, "srclib") &&
			!strings.Contains(step.Container.Image, "srclib-java") {
			srclibExplicitlyConfigured = true
			break
		}
	}

	usingSrclib := srclibExplicitlyConfigured // track if we've found any srclib languages
	unsupported := []string{}

	// Add the srclib build steps for all of the languages we
	// detect. But if we've explicitly configured srclib at all, then
	// don't do any automagic.
	if !srclibExplicitlyConfigured {
		for _, lang := range inv.Languages {
			b, ok := langSrclibConfigs[lang.Name]
			if !ok {
				unsupported = append(unsupported, lang.Name)
				continue
			}
			usingSrclib = true
			if err := insertSrclibBuild(config, axes, b); err != nil {
				return err
			}
		}
	}

	if len(unsupported) > 0 {
		b := buildLogMsg(
			fmt.Sprintf("Sourcegraph does not yet support %s", strings.Join(unsupported, ", ")),
			fmt.Sprintf("Sourcegraph does not yet support the following languages detected in this repository:\n%s\n", strings.Join(unsupported, "\n")),
		)
		if err := insertSrclibBuild(config, axes, b); err != nil {
			return err
		}
	}

	if !usingSrclib {
		if err := insertSrclibBuild(config, axes, buildLogMsg("Sourcegraph did not find any supported programming languages", "No supported programming languages were auto-detected.")); err != nil {
			return err
		}
	}

	// Insert the srclib import build step, only if we found an actual
	// srclib analyzer already (or one was explicitly configured).
	if srclibImportURL != nil && usingSrclib {
		if err := insertSrclibBuild(config, axes, srclibImportStep(srclibImportURL)); err != nil {
			return err
		}
	}

	return nil
}

// Note: If you push new Docker images for the srclib build steps, you
// MUST update the SHA256 digest, or else users will continue using
// the old Docker image. Also ensure you `docker push` the new Docker
// images, or else users' builds will fail because the required image
// is not available on the Docker Hub.
var langSrclibConfigs = map[string]droneyaml.BuildItem{
	"Go": {
		Key: "Go (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibGoImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"JavaScript": {
		Key: "JavaScript (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibJavaScriptImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"Java": {
		Key: "Java (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibJavaImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"TypeScript": {
		Key: "TypeScript (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibTypeScriptImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"C#": {
		Key: "C# (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibCSharpImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"CSS": {
		Key: "CSS (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibCSSImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	"Python": {
		Key: "Python (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: droneSrclibPythonImage,
			},
			Commands:     srclibBuildCommands,
			AllowFailure: true,
		},
	},
	// "HTML": {
	// 	Key: "HTML (indexing)",
	// 	Build: droneyaml.Build{
	// 		Container: droneyaml.Container{
	// 			Image: droneSrclibCSSImage,
	// 		},
	// 		Commands:     srclibBuildCommands,
	// 		AllowFailure: true,
	// 	},
	// },
}

var srclibBuildCommands = []string{"srclib config", "srclib make"}

// srclibImportStep returns a Drone build step that imports srclib
// data to the httpapi srclib import endpoint given by importURL
// (e.g., http://localhost:3080/.api/repos/my/repo/-/srclib-import).
func srclibImportStep(importURL *url.URL) droneyaml.BuildItem {
	return droneyaml.BuildItem{
		Key: "srclib import",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				// The hash is the final line of docker push output
				Image: "sourcegraph/srclib-import@sha256:1c3ae80a7fb7401e8a9b0ebea7ede2b21e25f7735cab64d9912e506216de93fe",
				Environment: droneyaml.MapEqualSlice([]string{
					"SOURCEGRAPH_IMPORT_URL=" + importURL.String(),
				}),
			},
			Commands: []string{
				"echo Importing to $SOURCEGRAPH_IMPORT_URL",
				`files=$(find .srclib-cache/ -type f | head -n 1); if [ -z "$files" ]; then echo No srclib data files found to import; exit 0; fi`,
				`cd .srclib-cache/* && /usr/bin/zip -q --no-dir-entries -r - . > /tmp/srclib-cache.zip`,
				`srclib-import $SOURCEGRAPH_IMPORT_URL /tmp/srclib-cache.zip`,
				"echo Done importing",
			},
		},
	}
}

// insertSrclibBuild inserts a build into the YAML. If there is a build
// matrix, the step will only execute for a single cell in the matrix.
func insertSrclibBuild(config *droneyaml.Config, axes []matrix.Axis, build droneyaml.BuildItem) error {
	if len(axes) < 1 {
		panic("must have at least 1 axis")
	}
	if len(axes) > 1 {
		build.Filter = droneyaml.Filter{Matrix: axes[0]}
	}

	{
		// If the build section is a single-build section, then we need to
		// convert it into a multi-build section and give the lone
		// existing build step a default name ("build").
		v, err := config.Build.MarshalYAML()
		if err != nil {
			return err
		}
		if v, ok := v.(droneyaml.Build); ok {
			// Is a single-section build.
			config.Build = droneyaml.Builds{{Key: "build", Build: v}}
		}
	}

	config.Build = append(config.Build, build)
	return nil
}
