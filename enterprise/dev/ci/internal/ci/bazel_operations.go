package ci

import (
	"fmt"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

func BazelOperations() *operations.Set {
	ops := operations.NewNamedSet("Bazel")
	ops.Append(bazelConfigure())
	ops.Append(bazelTest("//...", "//client/web:test"))
	ops.Append(bazelBackCompatTest(
		"@sourcegraph_back_compat//cmd/...",
		"@sourcegraph_back_compat//lib/...",
		"@sourcegraph_back_compat//internal/...",
		"@sourcegraph_back_compat//enterprise/cmd/...",
		"@sourcegraph_back_compat//enterprise/internal/...",
	))
	return ops
}

func bazelCmd(args ...string) string {
	pre := []string{
		"bazel",
		"--bazelrc=.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
	}
	rawCmd := append(pre, args...)
	return strings.Join(rawCmd, " ")
}

// bazelAnalysisPhase only runs the analasys phase, ensure that the buildfiles
// are correct, but do not actually build anything.
func bazelAnalysisPhase() func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmd := bazelCmd(
		"build",
		"--nobuild", // this is the key flag to enable this.
		"//...",
	)

	cmds := []bk.StepOpt{
		bk.Key("bazel-analysis"),
		bk.Agent("queue", "bazel"),
		bk.Cmd(cmd),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Analysis phase",
			cmds...,
		)
	}
}
func bazelConfigure() func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmds := []bk.StepOpt{
		bk.Key("bazel-configure"),
		bk.Agent("queue", "bazel"),
		bk.AnnotatedCmd("dev/ci/bazel-configure.sh", bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeWarning,
				IncludeNames: false,
			},
		}),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Ensure buildfiles are up to date",
			cmds...,
		)
	}
}

func bazelAnnouncef(format string, args ...any) bk.StepOpt {
	msg := fmt.Sprintf(format, args...)
	return bk.Cmd(fmt.Sprintf(`echo "--- :bazel: %s"`, msg))
}

func bazelTest(targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.DependsOn("bazel-configure"),
		bk.Agent("queue", "bazel"),
	}

	// Test commands
	bazelTestCmds := []bk.StepOpt{}
	for _, target := range targets {
		cmd := bazelCmd(fmt.Sprintf("test %s", target))
		bazelTestCmds = append(bazelTestCmds,
			bazelAnnouncef("bazel test %s", target),
			bk.Cmd(cmd))
	}
	cmds = append(cmds, bazelTestCmds...)

	// Run commands
	runTargets := []string{
		"//client/web:bundlesize-report",
	}
	bazelRunCmd := bazelCmd(fmt.Sprintf("run %s", strings.Join(runTargets, " ")))
	cmds = append(cmds,
		bazelAnnouncef("bazel run %s", strings.Join(runTargets, " ")),
		bk.Cmd(bazelRunCmd),
		bazelAnnouncef("✅"),
	)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Tests",
			cmds...,
		)
	}
}

func bazelBackCompatTest(targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.DependsOn("bazel-configure"),
		bk.Agent("queue", "bazel"),

		// Generate a patch that backports the migration from the new code into the old one.
		bk.Cmd("git diff origin/ci/backcompat-v5.0.0..HEAD -- migrations/ > dev/backcompat/patches/back_compat_migrations.patch"),
	}

	bazelRawCmd := bazelCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(
		cmds,
		bk.Cmd(bazelRawCmd),
	)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: BackCompat Tests",
			cmds...,
		)
	}
}

func bazelTestWithDepends(optional bool, dependsOn string, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Agent("queue", "bazel"),
	}

	bazelRawCmd := bazelCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.Cmd(bazelRawCmd))
	cmds = append(cmds, bk.DependsOn(dependsOn))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Tests",
			cmds...,
		)
	}
}

func bazelBuild(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Agent("queue", "bazel"),
	}
	bazelRawCmd := bazelCmd(fmt.Sprintf("build %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.Cmd(bazelRawCmd))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Build ...",
			cmds...,
		)
	}
}
