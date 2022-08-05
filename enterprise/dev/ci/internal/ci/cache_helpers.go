package ci

import "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"

func withYarnCache() buildkite.StepOpt {
	return buildkite.Cache(&buildkite.CacheOptions{
		ID:          "node_modules",
		Key:         "cache-node_modules-{{ checksum 'yarn.lock' }}",
		RestoreKeys: []string{"cache-node_modules-{{ checksum 'yarn.lock' }}"},
		Paths:       []string{"node_modules", "client/extension-api/node_modules"},
		// Compressing really slows down the process, as the node modules folder is huge. It's faster to just DL it.
		Compress: false,
	})
}
