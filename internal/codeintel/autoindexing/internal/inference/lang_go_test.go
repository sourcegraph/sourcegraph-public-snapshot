package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

const netrcString = `if [ "$NETRC_DATA" ]; then
  echo "Writing netrc config to $HOME/.netrc"
  echo "$NETRC_DATA" > ~/.netrc
else
  echo "No netrc config set, continuing"
fi
`

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
							Commands: []string{netrcString, "go mod download"},
						},
					},
					LocalSteps:       []string{netrcString},
					Root:             "foo/bar",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-go", "--no-animation"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"GOPRIVATE", "GOPROXY", "GONOPROXY", "GOSUMDB", "GONOSUMDB", "NETRC_DATA"},
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "foo/baz",
							Image:    expectedIndexerImage,
							Commands: []string{netrcString, "go mod download"},
						},
					},
					LocalSteps:       []string{netrcString},
					Root:             "foo/baz",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-go", "--no-animation"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"GOPRIVATE", "GOPROXY", "GONOPROXY", "GOSUMDB", "GONOSUMDB", "NETRC_DATA"},
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
					Steps:            nil,
					LocalSteps:       []string{netrcString},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"GO111MODULE=off", "scip-go", "--no-animation"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"GOPRIVATE", "GOPROXY", "GONOPROXY", "GOSUMDB", "GONOSUMDB", "NETRC_DATA"},
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
