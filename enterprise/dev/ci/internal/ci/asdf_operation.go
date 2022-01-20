package ci

import (
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
)

func ASDFInstall() []bk.StepOpt {
	return []bk.StepOpt{
		buildkite.Env("AWS_CONFIG_FILE", "/buildkite/.aws/config"),
		buildkite.Env("AWS_SHARED_CREDENTIALS_FILE", "/buildkite/.aws/credentials"),
		buildkite.Cache(&buildkite.CacheOptions{
			ID:          "asdf",
			Key:         "cache-asdf-{{ checksum '.tool-versions' }}",
			RestoreKeys: []string{"cache-asdf-{{ checksum '.tool-versions' }}"},
			Paths:       []string{"../../../../root/.asdf/installs"},
		}),
		buildkite.Cmd("asdf install"),
	}
}
