package ci

import (
	"fmt"

	"github.com/sourcegraph/log"

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
			bk.Env("DOCKER_BAZEL", "true"),
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

var reStepKeySanitizer = lazyregexp.New(`[^a-zA-Z0-9_-]+`)

// sanitizeStepKey sanitizes BuildKite StepKeys by removing any invalid characters
func sanitizeStepKey(key string) string {
	return reStepKeySanitizer.ReplaceAllString(key, "")
}
