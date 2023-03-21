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
	ops.Append(bazelConfigure())
	ops.Append(bazelBuildAndTest(optional, "//..."))
	return ops
}

func bazelConfigure() func(*bk.Pipeline) {
	configureCmd := []string{
		"bazel",
		"--bazelrc=.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
		"configure",
	}

	// if there are changes diff will exit with 1, and 0 otherwise
	gitDiff := "git diff --exit-code"
	cmds := []bk.StepOpt{
		bk.Key("bazel-configure"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
		bk.RawCmd(strings.Join(configureCmd, " ")),
		bk.RawCmd(gitDiff),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel: Configure",
			cmds...,
		)
	}
}

// bazelBuildAndTest will perform a build and test on the same agent, which is the preferred method
// over running them in two separate jobs, so we don't build the same code twice.
func bazelBuildAndTest(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.DependsOn("bazel-configure"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}

	for _, target := range targets {
		bazelBuildCmd := []string{
			"bazel",
			"--bazelrc=.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
			fmt.Sprintf("build %s", target),
			"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
			"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
		}

		bazelTestCmd := []string{
			"bazel",
			"--bazelrc=.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
			fmt.Sprintf("test %s", target),
			"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
			"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
		}
		cmds = append(
			cmds,
			bk.RawCmd(strings.Join(bazelBuildCmd, " ")),
			bk.RawCmd(strings.Join(bazelTestCmd, " ")),
		)
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}

		// TODO(JH) Broken we don't have go on the bazel agents
		// cmds = append(cmds, bk.SlackStepNotify(&bk.SlackStepNotifyConfigPayload{
		// 	Message:     ":alert: :bazel: test failed",
		// 	ChannelName: "dev-experience-alerts",
		// 	Conditions: bk.SlackStepNotifyPayloadConditions{
		// 		Failed:   true,
		// 		Branches: []string{"main"},
		// 	},
		// }))

		pipeline.AddStep(":bazel: Build && Test",
			cmds...,
		)
	}
}

func bazelTest(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}

	for _, target := range targets {
		bazelCmd := []string{
			"bazel",
			"--bazelrc=.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
			fmt.Sprintf("test %s", target),
			"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
			"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
		}
		cmds = append(cmds, bk.RawCmd(strings.Join(bazelCmd, " ")))
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}

		// TODO(JH) Broken we don't have go on the bazel agents
		// cmds = append(cmds, bk.SlackStepNotify(&bk.SlackStepNotifyConfigPayload{
		// 	Message:     ":alert: :bazel: test failed",
		// 	ChannelName: "dev-experience-alerts",
		// 	Conditions: bk.SlackStepNotifyPayloadConditions{
		// 		Failed:   true,
		// 		Branches: []string{"main"},
		// 	},
		// }))

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

	for _, target := range targets {
		bazelCmd := []string{
			"bazel",
			"--bazelrc=.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.bazelrc",
			"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
			fmt.Sprintf("build %s", target),
			"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
			"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
		}
		cmds = append(cmds, bk.RawCmd(strings.Join(bazelCmd, " ")))
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}

		// TODO(JH) Broken we don't have go on the bazel agents
		// cmds = append(cmds, bk.SlackStepNotify(&bk.SlackStepNotifyConfigPayload{
		// 	Message:     ":alert: :bazel: build failed",
		// 	ChannelName: "dev-experience-alerts",
		// 	Conditions: bk.SlackStepNotifyPayloadConditions{
		// 		Failed:   true,
		// 		Branches: []string{"main"},
		// 	},
		// }))

		pipeline.AddStep(":bazel: Build ...",
			cmds...,
		)
	}
}
