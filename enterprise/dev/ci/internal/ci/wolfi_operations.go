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

	for _, c := range changedFiles {
		match := packageRegex.FindStringSubmatch(c)
		if len(match) == 2 {
			ops.Append(buildPackages(match[1]))
		} else {
			logger.Fatal(fmt.Sprintf("Unable to extract package name from '%s', matches were %+v\n", c, match))
		}
	}

	return ops
}

func buildPackages(target string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":package: Package dependency '%s'", target),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-package.sh %s", target)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
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
