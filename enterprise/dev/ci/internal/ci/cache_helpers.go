package ci

import "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"

func withYarnCache() buildkite.StepOpt {
	return buildkite.Cache(&buildkite.CacheOptions{
		ID:          "node_modules",
		Key:         "cache-node_modules-{{ checksum 'yarn.lock' }}",
		RestoreKeys: []string{"cache-node_modules-{{ checksum 'yarn.lock' }}"},
		Paths:       []string{"node_modules", "client/extension-api/node_modules", "client/eslint-plugin-sourcegraph/node_modules"},
		// TODO: @jhchabran, check the numbers, but in my last run it seemed to be clear that compression is really slow (+3/4m IIRC)
		Compress: false,
	})
}
