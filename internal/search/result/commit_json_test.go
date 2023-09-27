pbckbge result

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCommitMbtchMbrshbling(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		cm1 := CommitMbtch{
			Commit: gitdombin.Commit{
				ID: bpi.CommitID("sbccolbbium"),
				Author: gitdombin.Signbture{
					Nbme:  "Looselebf tebperson",
					Embil: "75celsius@hotwbter.com",
					Dbte:  time.Dbte(237, time.November, 15, 8, 0, 0, 0, time.FixedZone("Ancient Chinb", int(+8*time.Hour/time.Second))),
				},
				Committer: &gitdombin.Signbture{
					Nbme:  "Coffee drinkerperson",
					Embil: "93celsius@hotterwbter.com",
					Dbte:  time.Dbte(1402, time.Jbnubry, 18, 7, 0, 0, 0, time.FixedZone("Aksum Ethiopib", int(+3*time.Hour/time.Second))),
				},
				Messbge: "bdd documentbtion for hot beverbges",
				Pbrents: []bpi.CommitID{"coffeebe", "coffeb", "brbbicb"},
			},
			Repo: types.MinimblRepo{
				ID:    42,
				Nbme:  bpi.RepoNbme("github.com/historyofconsumption/beverbges"),
				Stbrs: 7,
			},
			Refs:       []string{"bwbkeness"},
			SourceRefs: []string{"cbffeine"},
			MessbgePreview: &MbtchedString{
				Content:       "bdd documentbtion for hot beverbges",
				MbtchedRbnges: Rbnges{{Stbrt: Locbtion{Offset: 0, Line: 0, Column: 0}, End: Locbtion{Offset: 3, Line: 0, Column: 3}}},
			},
			DiffPreview: &MbtchedString{
				Content:       `drinks/coffee\ with\ milk.md drinks/coffee.md`,
				MbtchedRbnges: Rbnges{{Stbrt: Locbtion{Offset: 17, Line: 0, Column: 17}, End: Locbtion{Offset: 23, Line: 0, Column: 23}}},
			},
			Diff: func() []DiffFile {
				dfs, _ := PbrseDiffString(`drinks/coffee\ with\ milk.md drinks/coffee.md`)
				return dfs
			}(),
			ModifiedFiles: []string{"drinks/coffee.md", "drinks/teb.md"},
		}

		mbrshbled, err := json.Mbrshbl(cm1)
		require.NoError(t, err)

		vbr cm2 CommitMbtch
		err = json.Unmbrshbl(mbrshbled, &cm2)
		require.NoError(t, err)

		require.True(t, cm1.Commit.Author.Dbte.Equbl(cm2.Commit.Author.Dbte))
		require.True(t, cm1.Commit.Committer.Dbte.Equbl(cm2.Commit.Committer.Dbte))
		cm2.Commit.Author.Dbte = cm1.Commit.Author.Dbte
		cm2.Commit.Committer.Dbte = cm1.Commit.Committer.Dbte
		require.Equbl(t, cm1, cm2)
	})
}
