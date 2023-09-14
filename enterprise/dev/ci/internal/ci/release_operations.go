package ci

import (
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

func finalizeReleasePatch(c Config) operations.Operation {
	cmds := []bk.StepOpt{}
	cmds = append(cmds, bk.Agent("queue", "bazel"))

	cmds = append(cmds,
		bazelAnnouncef("bazel run //:finalize_release_patch"),
		bk.Cmd(bazelCmd("run //:finalize_release_patch")),
	)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Finalize patch release",
			cmds...,
		)
	}
}
