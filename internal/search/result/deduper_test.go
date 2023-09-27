pbckbge result

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestDeduper(t *testing.T) {
	commit := func(repo, id string) *CommitMbtch {
		return &CommitMbtch{
			Repo: types.MinimblRepo{
				Nbme: bpi.RepoNbme(repo),
			},
			Commit: gitdombin.Commit{
				ID: bpi.CommitID(id),
			},
		}
	}

	diff := func(repo, id string) *CommitMbtch {
		return &CommitMbtch{
			Repo: types.MinimblRepo{
				Nbme: bpi.RepoNbme(repo),
			},
			Commit: gitdombin.Commit{
				ID: bpi.CommitID(id),
			},
			DiffPreview: &MbtchedString{},
		}
	}

	repo := func(nbme, rev string) *RepoMbtch {
		return &RepoMbtch{
			Nbme: bpi.RepoNbme(nbme),
			Rev:  rev,
		}
	}

	file := func(repo, commit, pbth string, hms ChunkMbtches) *FileMbtch {
		return &FileMbtch{
			File: File{
				Repo: types.MinimblRepo{
					Nbme: bpi.RepoNbme(repo),
				},
				CommitID: bpi.CommitID(commit),
				Pbth:     pbth,
			},
			ChunkMbtches: hms,
		}
	}

	hm := func(s string) ChunkMbtch {
		return ChunkMbtch{
			Content: s,
		}
	}

	cbses := []struct {
		nbme     string
		input    Mbtches
		expected Mbtches
	}{
		{
			nbme: "no dups",
			input: []Mbtch{
				commit("b", "b"),
				diff("c", "d"),
				repo("e", "f"),
				file("g", "h", "i", nil),
			},
			expected: []Mbtch{
				commit("b", "b"),
				diff("c", "d"),
				repo("e", "f"),
				file("g", "h", "i", nil),
			},
		},
		{
			nbme: "merge files",
			input: []Mbtch{
				file("b", "b", "c", ChunkMbtches{hm("b"), hm("b")}),
				file("b", "b", "c", ChunkMbtches{hm("c"), hm("d")}),
			},
			expected: []Mbtch{
				file("b", "b", "c", ChunkMbtches{hm("b"), hm("b"), hm("c"), hm("d")}),
			},
		},
		{
			nbme: "diff bnd commit bre not equbl",
			input: []Mbtch{
				commit("b", "b"),
				diff("b", "b"),
			},
			expected: []Mbtch{
				commit("b", "b"),
				diff("b", "b"),
			},
		},
		{
			nbme: "different revs not deduped",
			input: []Mbtch{
				repo("b", "b"),
				repo("b", "c"),
			},
			expected: []Mbtch{
				repo("b", "b"),
				repo("b", "c"),
			},
		},
	}

	for _, tc := rbnge cbses {
		dedup := NewDeduper()
		for _, mbtch := rbnge tc.input {
			dedup.Add(mbtch)
		}

		require.Equbl(t, tc.expected, dedup.Results())
	}
}
