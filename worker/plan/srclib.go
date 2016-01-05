package plan

import (
	"fmt"
	"net/url"
	"strings"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
)

// configureSrclib modifies the Drone config to run srclib analysis
// during the CI build.
func configureSrclib(inv *inventory.Inventory, config *droneyaml.Config, axes []matrix.Axis, srclibImportURL *url.URL) error {
	var srclibExplicitlyConfigured bool
	for _, step := range config.Build {
		// Rough heuristic for now: does the Docker image name contain
		// "srclib".
		if strings.Contains(step.Container.Image, "srclib") {
			srclibExplicitlyConfigured = true
			break
		}
	}

	usingSrclib := srclibExplicitlyConfigured // track if we've found any srclib languages

	// Add the srclib build steps for all of the languages we
	// detect. But if we've explicitly configured srclib at all, then
	// don't do any automagic.
	if !srclibExplicitlyConfigured {
		for _, lang := range inv.Languages {
			b, ok := langSrclibConfigs[lang.Name]
			if ok {
				usingSrclib = true
			} else {
				b = buildLogMsg(fmt.Sprintf("Code Intelligence does not yet support %s", lang.Name), fmt.Sprintf("Sourcegraph Code Intelligence does not yet support %s (which was detected in this repository)", lang.Name))
			}

			if err := insertSrclibBuild(config, axes, b); err != nil {
				return err
			}
		}
	}

	if !usingSrclib {
		if err := insertSrclibBuild(config, axes, buildLogMsg("Code Intelligence did not find any supported programming languages", "no supported programming languages were auto-detected for Sourcegraph Code Intelligence")); err != nil {
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
	"Go": droneyaml.BuildItem{
		Key: "Go (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: "srclib/drone-srclib-go@sha256:3c7ea130a31efd438d7b93d89eda63085d656aea93efc26cd34b698c877cf413",
			},
			Commands: srclibBuildCommands,
		},
	},
	"JavaScript": droneyaml.BuildItem{
		Key: "JavaScript (indexing)",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: "srclib/drone-srclib-javascript@sha256:d02d302667c4cd8efa998b82ec95643c5924d774e247a77649aa7618d45a46cf",
			},
			Commands: srclibBuildCommands,
		},
	},
}

var srclibBuildCommands = []string{"srclib config", "srclib make"}

// srclibImportStep returns a Drone build step that imports srclib
// data to the httpapi srclib import endpoint given by importURL
// (e.g., http://localhost:3080/.api/repos/my/repo/.srclib-import).
func srclibImportStep(importURL *url.URL) droneyaml.BuildItem {
	return droneyaml.BuildItem{
		Key: "srclib import",
		Build: droneyaml.Build{
			Container: droneyaml.Container{
				Image: "library/alpine:3.2",
				Environment: droneyaml.MapEqualSlice{
					"SOURCEGRAPH_IMPORT_URL=" + importURL.String(),
				},
			},
			Commands: []string{
				"apk add -q -U ca-certificates zip curl",
				"rm -rf /var/cache/apk/*",
				"echo Importing to $SOURCEGRAPH_IMPORT_URL",
				`[ -d .srclib-cache/* ] || (echo No srclib data found to import; exit 1)`,
				`cd .srclib-cache/* && /usr/bin/zip -q --no-dir-entries -r - . | \
			 /usr/bin/curl \
				--silent --show-error \
				--fail \
				--netrc \
				--connect-timeout 3 \
				--max-time 300 \
				--no-keepalive \
				--retry 3 \
				--retry-delay 2 \
				-XPUT \
				-H 'Content-Type: application/x-zip-compressed' \
				-H 'Content-Transfer-Encoding: binary' \
				--data-binary @- \
				$SOURCEGRAPH_IMPORT_URL`,
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
