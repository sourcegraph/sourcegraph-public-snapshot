package ci

import "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"

func withPnpmCache() buildkite.StepOpt {
	return buildkite.Cache(&buildkite.CacheOptions{
		ID:          "node_modules_pnpm",
		Key:         "cache-node_modules-pnpm-{{ checksum 'pnpm-lock.yaml' }}",
		RestoreKeys: []string{"cache-node_modules-pnpm-{{ checksum 'pnpm-lock.yaml' }}"},
		Paths:       []string{"node_modules"},
		// Compressing really slows down the process, as the node modules folder is huge. It's faster to just DL it.
		Compress: false,
	})
}

func withBundleSizeCache(commit string) buildkite.StepOpt {
	return buildkite.Cache(&buildkite.CacheOptions{
		ID:          "bundle_size_cache",
		Key:         "bundle_size_cache-{{ git.commit }}",
		RestoreKeys: []string{"bundle_size_cache-{{ git.commit }}"},
		Paths:       []string{"client/web/dist/stats-" + commit + ".json"},
		Compress:    true,
	})
}
