package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const jsonTestInput = `
{
	"index_jobs": [
		{
			"steps": [
				{
					// Comments are the future
					"image": "go:latest",
					"commands": ["go mod vendor"],
				}
			],
			"indexer": "lsif-go",
			"indexer_args": ["--no-animation"],
		},
		{
			"root": "web/",
			"indexer": "scip-typescript",
			"indexer_args": ["index", "--yarn-workspaces"],
			"outfile": "lsif.dump",
		},
	]
}
`

func TestUnmarshalJSON(t *testing.T) {
	actual, err := UnmarshalJSON([]byte(jsonTestInput))
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

func TestJsonUnmarshal(t *testing.T) {
	const input = `
	{
		// comment
		/* another comment */
		"hello": "world",
	}`

	var actual any
	if err := jsonUnmarshal(input, &actual); err != nil {
		t.Fatalf("unexpected error unmarshalling payload: %s", err)
	}

	if diff := cmp.Diff(map[string]any{"hello": "world"}, actual); diff != "" {
		t.Errorf("unexpected configuration (-want +got):\n%s", diff)
	}
}
