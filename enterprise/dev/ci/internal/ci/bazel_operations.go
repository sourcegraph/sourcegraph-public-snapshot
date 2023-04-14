package ci

import (
	"fmt"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

const bazelRemoteCacheURL = "https://storage.googleapis.com/sourcegraph_bazel_cache"

func BazelOperations(optional bool) *operations.Set {
	ops := operations.NewNamedSet("Bazel")
	ops.Append(bazelConfigure(optional))
	ops.Append(bazelTest(optional, "//..."))
	return ops
}

func bazelRawCmd(args ...string) string {
	pre := []string{
		"bazel",
		"--bazelrc=.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
	}

	post := []string{
		"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
		"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
	}

	rawCmd := append(pre, args...)
	rawCmd = append(rawCmd, post...)
	return strings.Join(rawCmd, " ")
}

// bazelAnalysisPhase only runs the analasys phase, ensure that the buildfiles
// are correct, but do not actually build anything.
func bazelAnalysisPhase(optional bool) func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmd := bazelRawCmd(
		"build",
		"--nobuild", // this is the key flag to enable this.
		"//...",
	)

	cmds := []bk.StepOpt{
		bk.Key("bazel-analysis"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
		bk.RawCmd(cmd),
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}

		pipeline.AddStep(":bazel: Analysis phase",
			cmds...,
		)
	}
}
func bazelConfigure(optional bool) func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmds := []bk.StepOpt{
		bk.Key("bazel-configure"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
		bk.AnnotatedCmd("dev/ci/bazel-configure.sh", bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeWarning,
				IncludeNames: false,
			},
		}),
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Ensure buildfiles are up to date",
			cmds...,
		)
	}
}

func bazelTest(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.DependsOn("bazel-configure"),
		bk.Agent("queue", "bazel"),
	}

	bazelRawCmd := bazelRawCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(
		cmds,
		// TODO(JH): run server image for end-to-end tests on SOURCEGRAPH_BASE_URL similar to run-bazel-server.sh.
		bk.RawCmd(bazelRawCmd),
	)

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Tests",
			cmds...,
		)
	}
}

func bazelTestWithDepends(optional bool, dependsOn string, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}

	bazelRawCmd := bazelRawCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.RawCmd(bazelRawCmd))
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
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}
	bazelRawCmd := bazelRawCmd(fmt.Sprintf("build %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.RawCmd(bazelRawCmd))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Build ...",
			cmds...,
		)
	}
}
