package ci

import (
	"fmt"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
)

func bazelBuild(targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Key("bazel_build"),
		bk.Agent("queue", AspectWorkflows.QueueDefault),
	}
	cmd := bazelStampedCmd(fmt.Sprintf("build %s", strings.Join(targets, " ")))
	cmds = append(
		cmds,
		bk.Cmd(cmd),
		bk.Cmd(bazelStampedCmd("run //cmd/server:candidate_push")),
	)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Build ...",
			cmds...,
		)
	}
}

func bazelCmd(args ...string) string {
	genBazelRC, bazelrc := aspectBazelRC()
	pre := []string{
		genBazelRC,
		"bazel",
		fmt.Sprintf("--bazelrc=%s", bazelrc),
		fmt.Sprintf("--bazelrc=%s", ".aspect/bazelrc/ci.sourcegraph.bazelrc"),
	}
	Cmd := append(pre, args...)
	return strings.Join(Cmd, " ")
}

func bazelStampedCmd(args ...string) string {
	genBazelRC, bazelrc := aspectBazelRC()
	pre := []string{
		genBazelRC,
		"bazel",
		fmt.Sprintf("--bazelrc=%s", bazelrc),
		fmt.Sprintf("--bazelrc=%s", ".aspect/bazelrc/ci.sourcegraph.bazelrc"),
	}
	post := []string{
		"--stamp",
		"--workspace_status_command=./dev/bazel_stamp_vars.sh",
	}

	cmd := append(pre, args...)
	cmd = append(cmd, post...)
	return strings.Join(cmd, " ")
}

// TODO(burmudar): do we remove this?
func bazelPrechecks() func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Key("bazel-prechecks"),
		bk.SoftFail(100),
		bk.Agent("queue", AspectWorkflows.QueueDefault),
		bk.ArtifactPaths("./sg"),
		bk.AnnotatedCmd("dev/ci/bazel-prechecks.sh", bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeError,
				IncludeNames: false,
			},
		}),
		// We want to build sg on a bazel agent, but without the overhead
		// of its own pipeline step. After pre-checks have passed seems
		// the most natural, as we then know that the bazel files are
		// up-to-date for building sg.

		// TODO(burmudar): maybe move this to be part of gen pipeline?
		bk.Cmd("dev/ci/bazel-build-sg.sh"),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Bazel prechecks & build `sg`",
			cmds...,
		)
	}
}

func bazelAnnouncef(format string, args ...any) bk.StepOpt {
	msg := fmt.Sprintf(format, args...)
	return bk.Cmd(fmt.Sprintf(`echo "--- :bazel: %s"`, msg))
}
