package ci

import (
	"fmt"

	"github.com/sourcegraph/log"

	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var baseImageRegex = lazyregexp.New(`wolfi-images\/([^\/]+)\/\w+[.]yaml`)
var packageRegex = lazyregexp.New(`wolfi-packages\/(\w+)[.]yaml`)

func WolfiBaseImagesOperations(changedFiles []string) *operations.Set {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewSet()
	logger := log.Scoped("gen-pipeline", "generates the pipeline for ci")

	for _, c := range changedFiles {
		match := baseImageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			ops.Append(buildWolfi(match[1]))
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract base image name from '%s', matches were %+v\n", c, match))
		}
	}

	return ops
}

func WolfiPackagesOperations(changedFiles []string) *operations.Set {
	// TODO: Should we require the image name, or the full path to the yaml file?
	ops := operations.NewSet()
	logger := log.Scoped("gen-pipeline", "generates the pipeline for ci")

	var stepKeys []string
	for _, c := range changedFiles {
		match := packageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			buildFunc, key := buildPackages(match[1])
			stepKeys = append(stepKeys, key)
			ops.Append(buildFunc)
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract package name from '%s', matches were %+v\n", c, match))
		}
	}

	ops.Append(buildRepoIndex("main", stepKeys))

	return ops
}

func buildPackages(target string) (func(*bk.Pipeline), string) {
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
		pipeline.AddStep(fmt.Sprintf(":card_index_dividers: Building and signing repository index for branch '%s'", branch),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-repo-index.sh %s", branch)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
			// Depend on all previous package building steps
			bk.DependsOn(packageKeys...),
		)
	}
}

func buildWolfi(target string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":wolf: Build Wolfi base image '%s'", target),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-base-image.sh %s", target)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
		)
	}
}
