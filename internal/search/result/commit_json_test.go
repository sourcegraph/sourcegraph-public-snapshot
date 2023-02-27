package result

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCommitMatchMarshaling(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		cm1 := CommitMatch{
			Commit: gitdomain.Commit{
				ID: api.CommitID("saccolabium"),
				Author: gitdomain.Signature{
					Name:  "Looseleaf teaperson",
					Email: "75celsius@hotwater.com",
					Date:  time.Date(237, time.November, 15, 8, 0, 0, 0, time.FixedZone("Ancient China", int(+8*time.Hour/time.Second))),
				},
				Committer: &gitdomain.Signature{
					Name:  "Coffee drinkerperson",
					Email: "93celsius@hotterwater.com",
					Date:  time.Date(1402, time.January, 18, 7, 0, 0, 0, time.FixedZone("Aksum Ethiopia", int(+3*time.Hour/time.Second))),
				},
				Message: "add documentation for hot beverages",
				Parents: []api.CommitID{"coffeeae", "coffea", "arabica"},
			},
			Repo: types.MinimalRepo{
				ID:    42,
				Name:  api.RepoName("github.com/historyofconsumption/beverages"),
				Stars: 7,
			},
			Refs:       []string{"awakeness"},
			SourceRefs: []string{"caffeine"},
			MessagePreview: &MatchedString{
				Content:       "add documentation for hot beverages",
				MatchedRanges: Ranges{{Start: Location{Offset: 0, Line: 0, Column: 0}, End: Location{Offset: 3, Line: 0, Column: 3}}},
			},
			DiffPreview: &MatchedString{
				Content:       `drinks/coffee\ with\ milk.md drinks/coffee.md`,
				MatchedRanges: Ranges{{Start: Location{Offset: 17, Line: 0, Column: 17}, End: Location{Offset: 23, Line: 0, Column: 23}}},
			},
			Diff: func() []DiffFile {
				dfs, _ := ParseDiffString(`drinks/coffee\ with\ milk.md drinks/coffee.md`)
				return dfs
			}(),
			ModifiedFiles: []string{"drinks/coffee.md", "drinks/tea.md"},
		}

		marshaled, err := json.Marshal(cm1)
		require.NoError(t, err)

		var cm2 CommitMatch
		err = json.Unmarshal(marshaled, &cm2)
		require.NoError(t, err)

		require.True(t, cm1.Commit.Author.Date.Equal(cm2.Commit.Author.Date))
		require.True(t, cm1.Commit.Committer.Date.Equal(cm2.Commit.Committer.Date))
		cm2.Commit.Author.Date = cm1.Commit.Author.Date
		cm2.Commit.Committer.Date = cm1.Commit.Committer.Date
		require.Equal(t, cm1, cm2)
	})
}
