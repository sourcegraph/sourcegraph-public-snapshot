package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const yamlTestInput = `
index_jobs:
  -
    steps:
      - image: go:latest
        commands:
          - go mod vendor
    indexer: lsif-go
    indexer_args:
      - --no-animation
  -
    root: web/
    indexer: scip-typescript
    indexer_args: ['index', '--yarn-workspaces']
    outfile: lsif.dump
`

func TestUnmarshalYAML(t *testing.T) {
	actual, err := UnmarshalYAML([]byte(yamlTestInput))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := AutoIndexJobSpecList{
		JobSpecs: []AutoIndexJobSpec{
			{
				Steps: []DockerStep{
					{
						Root:     "",
						Image:    "go:latest",
						Commands: []string{"go mod vendor"},
					},
				},
				Indexer:     "lsif-go",
				IndexerArgs: []string{"--no-animation"},
			},
			{
				Steps:       nil,
				Root:        "web/",
				Indexer:     "scip-typescript",
				IndexerArgs: []string{"index", "--yarn-workspaces"},
				Outfile:     "lsif.dump",
			},
		},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected configuration (-want +got):\n%s", diff)
	}
}
