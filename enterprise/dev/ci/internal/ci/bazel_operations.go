package ci

import (
	"fmt"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

const bazelRemoteCacheURL = "https://storage.googleapis.com/sourcegraph_bazel_cache"

func BazelOperations() *operations.Set {
	ops := operations.NewSet()
	ops.Append(build("//dev/sg"))
	return ops
}

func build(target string) func(*bk.Pipeline) {
	bazelCmd := []string{
		"bazel",
		"--bazelrc=.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.bazelrc",
		fmt.Sprintf("build %s", target),
		"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
		"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":bazel: Build %s", target),
			bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
			bk.Cmd(strings.Join(bazelCmd, " ")),
			bk.Agent("queue", "bazel"),
		)
	}
}
