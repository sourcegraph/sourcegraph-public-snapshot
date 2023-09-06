package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRubyGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("ruby")

	testGenerators(t,
		generatorTestCase{
			description: "scip-ruby",
			repositoryContents: map[string]string{
				"a/Gemfile":               "",
				"a/a.gemspec":             "",
				"a/Gemfile.lock":          "",
				"a/rubygems-metadata.yml": "",
				"b/Gemfile":               "",
				"c/Gemfile.lock":          "",
				"d/d.gemspec":             "",
				"e/rubygems-metadata.yml": "",
			},
			expected: func() []config.IndexJob {
				var out []config.IndexJob
				dirs := []string{"a", "b", "c", "d", "e"}
				for _, dir := range dirs {
					out = append(out, config.IndexJob{
						Steps:       nil,
						LocalSteps:  nil,
						Root:        dir,
						Indexer:     expectedIndexerImage,
						IndexerArgs: []string{"scip-ruby-autoindex"},
						Outfile:     "index.scip",
					})
				}
				return out
			}(),
		},
	)
}
