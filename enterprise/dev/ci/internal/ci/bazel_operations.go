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
	ops.Append(bazelBuild(optional, "//..."))
	ops.Append(bazelTest(optional, "//..."))
	return ops
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
