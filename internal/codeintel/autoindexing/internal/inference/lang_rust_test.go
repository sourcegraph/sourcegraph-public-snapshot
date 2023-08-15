package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRustGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("rust")

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
					IndexerArgs: []string{"scip-rust", "index"},
					Outfile:     "index.scip",
				},
			},
		},
	)
}
