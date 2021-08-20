package result

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestDeduper(t *testing.T) {
	commit := func(repo, id string) *CommitMatch {
		return &CommitMatch{
			Repo: types.RepoName{
				Name: api.RepoName(repo),
			},
			Commit: git.Commit{
				ID: api.CommitID(id),
			},
		}
	}

	diff := func(repo, id string) *CommitMatch {
		return &CommitMatch{
			Repo: types.RepoName{
				Name: api.RepoName(repo),
			},
			Commit: git.Commit{
				ID: api.CommitID(id),
			},
			DiffPreview: &HighlightedString{},
		}
	}

	repo := func(name, rev string) *RepoMatch {
		return &RepoMatch{
			Name: api.RepoName(name),
			Rev:  rev,
		}
	}

	file := func(repo, commit, path string, lines []*LineMatch) *FileMatch {
		return &FileMatch{
			File: File{
				Repo: types.RepoName{
					Name: api.RepoName(repo),
				},
				CommitID: api.CommitID(commit),
				Path:     path,
			},
			LineMatches: lines,
		}
	}

	lm := func(s string) *LineMatch {
		return &LineMatch{
			Preview: s,
		}
	}

	cases := []struct {
		name     string
		input    []Match
		expected []Match
	}{
		{
			name: "no dups",
			input: []Match{
				commit("a", "b"),
				diff("c", "d"),
				repo("e", "f"),
				file("g", "h", "i", nil),
			},
			expected: []Match{
				commit("a", "b"),
				diff("c", "d"),
				repo("e", "f"),
				file("g", "h", "i", nil),
			},
		},
		{
			name: "merge files",
			input: []Match{
				file("a", "b", "c", []*LineMatch{lm("a"), lm("b")}),
				file("a", "b", "c", []*LineMatch{lm("c"), lm("d")}),
			},
			expected: []Match{
				file("a", "b", "c", []*LineMatch{lm("a"), lm("b"), lm("c"), lm("d")}),
			},
		},
		{
			name: "diff and commit are not equal",
			input: []Match{
				commit("a", "b"),
				diff("a", "b"),
			},
			expected: []Match{
				commit("a", "b"),
				diff("a", "b"),
			},
		},
		{
			name: "different revs not deduped",
			input: []Match{
				repo("a", "b"),
				repo("a", "c"),
			},
			expected: []Match{
				repo("a", "b"),
				repo("a", "c"),
			},
		},
	}

	for _, tc := range cases {
		dedup := NewDeduper()
		for _, match := range tc.input {
			dedup.Add(match)
		}

		require.Equal(t, tc.expected, dedup.Results())
	}
}
