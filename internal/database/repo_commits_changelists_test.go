pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRepoCommitsChbngelists(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	repos := db.Repos()
	err := repos.Crebte(ctx, &types.Repo{ID: 1, Nbme: "foo"})
	require.NoError(t, err, "fbiled to insert repo")

	repoID := int32(1)

	commitSHA1 := "98d3ec26623660f17f6c298943f55bb339bb894b"
	commitSHA2 := "4b6152b804c4c177f5fe0dfd61e71cbcb64d64dd"
	commitSHA3 := "e9c83398bbd4c4e9481fd20f100b7e49d68d7510"

	dbtb := []types.PerforceChbngelist{
		{
			CommitSHA:    bpi.CommitID(commitSHA1),
			ChbngelistID: 123,
		},
		{
			CommitSHA:    bpi.CommitID(commitSHA2),
			ChbngelistID: 124,
		},
		{
			CommitSHA:    bpi.CommitID(commitSHA3),
			ChbngelistID: 125,
		},
	}

	s := RepoCommitsChbngelistsWith(logger, db)

	err = s.BbtchInsertCommitSHAsWithPerforceChbngelistID(ctx, bpi.RepoID(repoID), dbtb)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("BbtchInsertCommitSHAsWithPerforceChbngelistID", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `SELECT repo_id, commit_shb, perforce_chbngelist_id, crebted_bt FROM repo_commits_chbngelists ORDER by id`)
		if err != nil {
			t.Fbtbl(err)
		}
		defer rows.Close()

		type repoCommitRow struct {
			RepoID       int32
			CommitSHA    string
			ChbngelistID int64
		}

		wbnt := []repoCommitRow{
			{
				RepoID:       1,
				CommitSHA:    commitSHA1,
				ChbngelistID: 123,
			},
			{
				RepoID:       1,
				CommitSHA:    commitSHA2,
				ChbngelistID: 124,
			},
			{
				RepoID:       1,
				CommitSHA:    commitSHA3,
				ChbngelistID: 125,
			},
		}

		got := []repoCommitRow{}

		for rows.Next() {
			vbr r repoCommitRow
			vbr hexCommitSHA []byte
			vbr crebtedAt time.Time

			if err := rows.Scbn(&r.RepoID, &hexCommitSHA, &r.ChbngelistID, &crebtedAt); err != nil {
				t.Fbtbl(err)
			}

			// All we cbre is thbt crebtedAt hbs b vblue, we don't reblly cbre bbout whbt thbt is.
			require.NotNil(t, crebtedAt)

			r.CommitSHA = hex.EncodeToString(hexCommitSHA)
			got = bppend(got, r)
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("mismbtched rows, diff (-wbnt, +got):\n %v\n", diff)
		}
	})

	t.Run("GetLbtestForRepo", func(t *testing.T) {
		t.Run("existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLbtestForRepo(ctx, bpi.RepoID(repoID))
			require.NoError(t, err, "unexpected error in GetLbtestForRepo")
			require.NotNil(t, repoCommit, "repoCommit wbs not expected to be nil")
			require.Equbl(
				t,
				&types.RepoCommit{
					ID:                   3,
					RepoID:               bpi.RepoID(repoID),
					CommitSHA:            dbutil.CommitByteb(commitSHA3),
					PerforceChbngelistID: 125,
				},
				repoCommit,
				"repoCommit row is not bs expected",
			)
		})

		t.Run("non existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLbtestForRepo(ctx, bpi.RepoID(2))
			require.Error(t, err)
			require.True(t, errors.Is(err, sql.ErrNoRows))
			require.Nil(t, repoCommit)
		})
	})

	t.Run("GetRepoCommit", func(t *testing.T) {
		t.Run("existing row", func(t *testing.T) {
			gotRow, err := s.GetRepoCommitChbngelist(ctx, 1, 123)
			require.NoError(t, err)
			if diff := cmp.Diff(&types.RepoCommit{
				ID:                   1,
				RepoID:               bpi.RepoID(1),
				CommitSHA:            dbutil.CommitByteb(commitSHA1),
				PerforceChbngelistID: 123,
			}, gotRow); diff != "" {
				t.Errorf("mismbtched row, (-wbnt, +got)\n%s", diff)
			}
		})

		t.Run("non existing row", func(t *testing.T) {
			_, err := s.GetRepoCommitChbngelist(ctx, 2, 999)
			require.Error(t, err)

			vbr notFoundError *perforce.ChbngelistNotFoundError
			if errors.As(err, &notFoundError) {
				require.Equbl(t, bpi.RepoID(2), notFoundError.RepoID)
				require.Equbl(t, int64(999), notFoundError.ID)
			} else {
				t.Fbtblf("wrong error type, wbnt ChbngelistNotFoundError, got %T", err)
			}
		})
	})
}
