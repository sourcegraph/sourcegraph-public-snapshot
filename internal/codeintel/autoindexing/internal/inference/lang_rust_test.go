package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRustGenerator(t *testing.T) {
	expectedIndexerImage := "sourcegraph/lsif-rust@sha256:83cb769788987eb52f21a18b62d51ebb67c9436e1b0d2e99904c70fef424f9d1"

	testGenerators(t,
		generatorTestCase{
			description: "rust-analyzer",
			repositoryContents: map[string]string{
				"foo/bar/Cargo.toml": "",
				"foo/baz/Cargo.toml": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     expectedIndexerImage,
					IndexerArgs: []string{"lsif-rust", "index"},
					Outfile:     "dump.lsif",
				},
			},
		},
	)
}
