package ci

import (
	"fmt"

	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

func WolfiOperations() *operations.Set {
	ops := operations.NewSet()
	ops.Append(buildPackages("foobar"))
	ops.Append(buildWolfi("foobar"))
	return ops
}

func buildPackages(target string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":package: Package dependency %s", target),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-package.sh %s", target)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
		)
	}
}

func buildWolfi(target string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":wolf: Build stuff %s", target),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build.sh %s", target)),
			// We want to run on the bazel queue, so we have a pretty minimal agent.
			bk.Agent("queue", "bazel"),
		)
	}
}
