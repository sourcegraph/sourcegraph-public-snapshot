package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var baseImageRegex = lazyregexp.New(`wolfi-images\/([\w-]+)[.]yaml`)
var packageRegex = lazyregexp.New(`wolfi-packages\/([\w-]+)[.]yaml`)

// WolfiPackagesOperations rebuilds any packages whose configurations have changed
func WolfiPackagesOperations(changedFiles []string) (*operations.Set, int) {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewNamedSet("Dependency packages")

	var buildStepKeys []string
	for _, c := range changedFiles {
		match := packageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			buildFunc, key := buildPackage(match[1])
			ops.Append(buildFunc)
			buildStepKeys = append(buildStepKeys, key)
		}
	}

	ops.Append(buildRepoIndex("main", buildStepKeys))

	return ops, len(buildStepKeys)
}

// WolfiBaseImagesOperations rebuilds any base images whose configurations have changed
func WolfiBaseImagesOperations(changedFiles []string, tag string, packagesChanged bool) (*operations.Set, int) {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewNamedSet("Base image builds")
	logger := log.Scoped("gen-pipeline", "generates the pipeline for ci")

	var buildStepKeys []string
	for _, c := range changedFiles {
		match := baseImageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			buildFunc, key := buildWolfiBaseImage(match[1], tag, packagesChanged)
			ops.Append(buildFunc)
			buildStepKeys = append(buildStepKeys, key)
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract base image name from '%s', matches were %+v\n", c, match))
		}
	}

	ops.Append(allBaseImagesBuilt(buildStepKeys))

	return ops, len(buildStepKeys)
}

// WolfiImagesOperations builds the specified docker images, or all images if none are provided
func WolfiImagesOperations(buildImages []string, version string, tag string, baseImagesChanged bool) *operations.Set {
	// If buildImages is not specified, rebuild all images
	// TODO: Maintain a list of Wolfi-based images?
	if len(buildImages) == 0 {
		buildImages = images.SourcegraphDockerImages
	}

	wolfiImageBuildOps := operations.NewNamedSet("Wolfi image builds")

	for _, dockerImage := range buildImages {
		// Don't upload sourcemaps
		// wolfiImageBuildOps.Append(buildCandidateDockerImage(dockerImage, version, tag, false))
		wolfiImageBuildOps.Append(
			buildCandidateWolfiDockerImage(dockerImage, version, tag, false, baseImagesChanged),
		)
	}

	return wolfiImageBuildOps
}

// Dependency tree between steps:
// (buildPackage[1], buildPackage[2], ...) <-- buildRepoIndex <-- (buildWolfi[1], buildWolfi[2], ...)

func buildPackage(target string) (func(*bk.Pipeline), string) {
	stepKey := sanitizeStepKey(fmt.Sprintf("package-dependency-%s", target))

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":package: Package dependency '%s'", target),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-package.sh %s", target)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
			bk.Key(stepKey),
		)
	}, stepKey
}

func buildRepoIndex(branch string, packageKeys []string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":card_index_dividers: Build and sign repository index for branch '%s'", branch),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-repo-index.sh %s", branch)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
			// Depend on all previous package building steps
			bk.DependsOn(packageKeys...),
			bk.Key("buildRepoIndex"),
		)
	}
}

func buildWolfiBaseImage(target string, tag string, dependOnPackages bool) (func(*bk.Pipeline), string) {
	stepKey := sanitizeStepKey(fmt.Sprintf("build-base-image-%s", target))

	return func(pipeline *bk.Pipeline) {

		opts := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-base-image.sh %s %s", target, tag)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
			bk.Key(stepKey),
		}
		// If packages have changed, wait for repo to be re-indexed as base images may depend on new packages
		if dependOnPackages {
			opts = append(opts, bk.DependsOn("buildRepoIndex"))
		}

		pipeline.AddStep(
			fmt.Sprintf(":octopus: Build Wolfi base image '%s'", target),
			opts...,
		)
	}, stepKey
}

// No-op to ensure all base images are updated before building full images
func allBaseImagesBuilt(baseImageKeys []string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":octopus: All base images built",
			bk.Cmd("echo 'All base images built'"),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
			// Depend on all previous package building steps
			bk.DependsOn(baseImageKeys...),
			bk.Key("buildAllBaseImages"),
		)
	}
}

// Build a candidate Wolfi docker image
func buildCandidateWolfiDockerImage(app, version, tag string, uploadSourcemaps bool, hasDependency bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		image := strings.ReplaceAll(app, "/", "-")
		localImage := "sourcegraph/wolfi-" + image + ":" + version

		cmds := []bk.StepOpt{
			bk.Key(candidateImageStepKey(app)),
			bk.Cmd(fmt.Sprintf(`echo "Building Wolfi %s image..."`, app)),
			bk.Env("DOCKER_BUILDKIT", "1"),
			bk.Env("IMAGE", localImage),
			bk.Env("VERSION", version),
			bk.Agent("queue", "bazel"),
		}

		if hasDependency {
			cmds = append(cmds, bk.DependsOn("buildAllBaseImages"))
		}

		// Add Sentry environment variables if we are building off main branch
		// to enable building the webapp with source maps enabled
		if uploadSourcemaps {
			cmds = append(cmds,
				bk.Env("SENTRY_UPLOAD_SOURCE_MAPS", "1"),
				bk.Env("SENTRY_ORGANIZATION", "sourcegraph"),
				bk.Env("SENTRY_PROJECT", "sourcegraph-dot-com"),
			)
		}

		// Allow all build scripts to emit info annotations
		buildAnnotationOptions := bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeInfo,
				IncludeNames: true,
			},
		}

		if _, err := os.Stat(filepath.Join("docker-images", app)); err == nil {
			// Building Docker image located under $REPO_ROOT/docker-images/
			cmds = append(cmds,
				bk.Cmd("ls -lah "+filepath.Join("docker-images", app, "build-wolfi.sh")),
				bk.Cmd(filepath.Join("docker-images", app, "build-wolfi.sh")))
		} else {
			// Building Docker images located under $REPO_ROOT/cmd/
			cmdDir := func() string {
				folder := app
				if app == "blobstore2" {
					// experiment: cmd/blobstore is a Go rewrite of docker-images/blobstore. While
					// it is incomplete, we do not want cmd/blobstore/Dockerfile to get publishe
					// under the same name.
					// https://github.com/sourcegraph/sourcegraph/issues/45594
					// TODO(blobstore): remove this when making Go blobstore the default
					folder = "blobstore"
				}
				// If /enterprise/cmd/... does not exist, build just /cmd/... instead.
				if _, err := os.Stat(filepath.Join("enterprise/cmd", folder)); err != nil {
					return "cmd/" + folder
				}
				return "enterprise/cmd/" + folder
			}()
			preBuildScript := cmdDir + "/pre-build.sh"
			if _, err := os.Stat(preBuildScript); err == nil {
				// Allow all
				cmds = append(cmds, bk.AnnotatedCmd(preBuildScript, buildAnnotationOptions))
			}
			cmds = append(cmds, bk.AnnotatedCmd(cmdDir+"/build-wolfi.sh", buildAnnotationOptions))
		}

		// Add "wolfi" to image name so we don't overwrite Alpine dev images
		wolfiApp := fmt.Sprintf("wolfi-%s", app)
		devImage := images.DevRegistryImage(wolfiApp, tag)
		cmds = append(cmds,
			// Retag the local image for dev registry
			bk.Cmd(fmt.Sprintf("docker tag %s %s", localImage, devImage)),
			// Publish tagged image
			bk.Cmd(fmt.Sprintf("docker push %s || exit 10", devImage)),
			// Retry in case of flakes when pushing
			bk.AutomaticRetryStatus(3, 10),
			// Retry in case of flakes when pushing
			bk.AutomaticRetryStatus(3, 222),
		)

		pipeline.AddStep(fmt.Sprintf(":octopus: :docker: :construction: Build Wolfi-based %s", app), cmds...)
	}
}

var reStepKeySanitizer = lazyregexp.New(`[^a-zA-Z0-9_-]+`)

// sanitizeStepKey sanitizes BuildKite StepKeys by removing any invalid characters
func sanitizeStepKey(key string) string {
	return reStepKeySanitizer.ReplaceAllString(key, "")
}
