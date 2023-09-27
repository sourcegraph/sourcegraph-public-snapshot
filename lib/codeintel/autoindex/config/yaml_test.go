pbckbge config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const ybmlTestInput = `
index_jobs:
  -
    steps:
      - imbge: go:lbtest
        commbnds:
          - go mod vendor
    indexer: lsif-go
    indexer_brgs:
      - --no-bnimbtion
  -
    root: web/
    indexer: scip-typescript
    indexer_brgs: ['index', '--ybrn-workspbces']
    outfile: lsif.dump
`

func TestUnmbrshblYAML(t *testing.T) {
	bctubl, err := UnmbrshblYAML([]byte(ybmlTestInput))
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	expected := IndexConfigurbtion{
		IndexJobs: []IndexJob{
			{
				Steps: []DockerStep{
					{
						Root:     "",
						Imbge:    "go:lbtest",
						Commbnds: []string{"go mod vendor"},
					},
				},
				Indexer:     "lsif-go",
				IndexerArgs: []string{"--no-bnimbtion"},
			},
			{
				Steps:       nil,
				Root:        "web/",
				Indexer:     "scip-typescript",
				IndexerArgs: []string{"index", "--ybrn-workspbces"},
				Outfile:     "lsif.dump",
			},
		},
	}
	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected configurbtion (-wbnt +got):\n%s", diff)
	}
}
