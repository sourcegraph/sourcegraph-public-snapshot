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

func WolfiBaseImagesOperations(changedFiles []string, tag string, packagesChanged bool) *operations.Set {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewNamedSet("Base image builds")
	logger := log.Scoped("gen-pipeline", "generates the pipeline for ci")

	for _, c := range changedFiles {
		match := baseImageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			ops.Append(buildWolfi(match[1], tag, packagesChanged))
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract base image name from '%s', matches were %+v\n", c, match))
		}
	}

	return ops
}

func WolfiPackagesOperations(changedFiles []string) *operations.Set {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewNamedSet("Dependency packages")
	logger := log.Scoped("gen-pipeline", "generates the pipeline for ci")

	var stepKeys []string
	for _, c := range changedFiles {
		match := packageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			buildFunc, key := buildPackage(match[1])
			stepKeys = append(stepKeys, key)
			ops.Append(buildFunc)
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract package name from '%s', matches were %+v\n", c, match))
		}
	}

	ops.Append(buildRepoIndex("main", stepKeys))

	return ops
}

// Dependency tree between steps:
// (buildPackage[1], buildPackage[2], ...) <-- buildRepoIndex <-- (buildWolfi[1], buildWolfi[2], ...)

func buildPackage(target string) (func(*bk.Pipeline), string) {
	// TODO: Can this be sanitised?
	stepKey := fmt.Sprintf("package-dependency-%s", target)

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

func buildWolfi(target string, tag string, dependOnPackages bool) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {

		opts := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-base-image.sh %s %s", target, tag)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
		}
		// If packages have changed, wait for repo to be re-indexed as base images may depend on new packages
		if dependOnPackages {
			opts = append(opts, bk.DependsOn("buildRepoIndex"))
		}

		pipeline.AddStep(
			fmt.Sprintf(":octopus: Build Wolfi base image '%s'", target),
			opts...,
		)
	}
}
