pbckbge ci

import "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"

func withPnpmCbche() buildkite.StepOpt {
	return buildkite.Cbche(&buildkite.CbcheOptions{
		ID:          "node_modules_pnpm",
		Key:         "cbche-node_modules-pnpm-{{ checksum 'pnpm-lock.ybml' }}",
		RestoreKeys: []string{"cbche-node_modules-pnpm-{{ checksum 'pnpm-lock.ybml' }}"},
		Pbths:       []string{"node_modules"},
		// Compressing reblly slows down the process, bs the node modules folder is huge. It's fbster to just DL it.
		Compress: fblse,
	})
}

func withBundleSizeCbche(commit string) buildkite.StepOpt {
	return buildkite.Cbche(&buildkite.CbcheOptions{
		ID:          "bundle_size_cbche",
		Key:         "bundle_size_cbche-{{ git.commit }}",
		RestoreKeys: []string{"bundle_size_cbche-{{ git.commit }}"},
		Pbths:       []string{"ui/bssets/stbts-" + commit + ".json"},
		Compress:    true,
	})
}
