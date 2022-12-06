package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestGoGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("go")

	testGenerators(t,
		generatorTestCase{
			description: "go modules",
			repositoryContents: map[string]string{
				"foo/bar/go.mod": "",
				"foo/baz/go.mod": "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "foo/bar",
							Image:    expectedIndexerImage,
							Commands: []string{"go mod download"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/bar",
					Indexer:     expectedIndexerImage,
					IndexerArgs: []string{"lsif-go", "--no-animation"},
					Outfile:     "",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "foo/baz",
							Image:    expectedIndexerImage,
							Commands: []string{"go mod download"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/baz",
					Indexer:     expectedIndexerImage,
					IndexerArgs: []string{"lsif-go", "--no-animation"},
					Outfile:     "",
				},
			},
		},
		generatorTestCase{
			description: "go files in root",
			repositoryContents: map[string]string{
				"main.go":       "",
				"internal/a.go": "",
				"internal/b.go": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     expectedIndexerImage,
					IndexerArgs: []string{"GO111MODULE=off", "lsif-go", "--no-animation"},
					Outfile:     "",
				},
			},
		},
		generatorTestCase{
			description: "go files in non-root (no match)",
			repositoryContents: map[string]string{
				"cmd/src/main.go": "",
			},
			expected: []config.IndexJob{},
		},
	)
}
