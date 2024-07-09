package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRepoCommitsChangelists(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	repos := db.Repos()
	err := repos.Create(ctx, &types.Repo{ID: 1, Name: "foo"})
	require.NoError(t, err, "failed to insert repo")
	err = repos.Create(ctx, &types.Repo{ID: 2, Name: "bar"})
	require.NoError(t, err, "failed to insert repo")

	repoID1 := int32(1)
	repoID2 := int32(2)

	commitSHA1 := "98d3ec26623660f17f6c298943f55aa339aa894a"
	commitSHA2 := "4b6152a804c4c177f5fe0dfd61e71cacb64d64dd"
	commitSHA3 := "e9c83398bbd4c4e9481fd20f100a7e49d68d7510"
	commitSHA4 := "aac83398bbd4c4e9481fd20f100a7e49d68d7510"

	dataForRepo1 := []types.PerforceChangelist{
		{
			CommitSHA:    api.CommitID(commitSHA1),
			ChangelistID: 123,
		},
		{
			CommitSHA:    api.CommitID(commitSHA2),
			ChangelistID: 124,
		},
		{
			CommitSHA:    api.CommitID(commitSHA3),
			ChangelistID: 125,
		},
	}
	dataForRepo2 := []types.PerforceChangelist{
		{
			CommitSHA:    api.CommitID(commitSHA4),
			ChangelistID: 126,
		},
	}
	s := RepoCommitsChangelistsWith(logger, db)

	err = s.BatchInsertCommitSHAsWithPerforceChangelistID(ctx, api.RepoID(repoID1), dataForRepo1)
	if err != nil {
		t.Fatal(err)
	}
	err = s.BatchInsertCommitSHAsWithPerforceChangelistID(ctx, api.RepoID(repoID2), dataForRepo2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("BatchInsertCommitSHAsWithPerforceChangelistID", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `SELECT repo_id, commit_sha, perforce_changelist_id, created_at FROM repo_commits_changelists WHERE repo_id=1 ORDER by id`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		type repoCommitRow struct {
			RepoID       int32
			CommitSHA    string
			ChangelistID int64
		}

		want := []repoCommitRow{
			{
				RepoID:       1,
				CommitSHA:    commitSHA1,
				ChangelistID: 123,
			},
			{
				RepoID:       1,
				CommitSHA:    commitSHA2,
				ChangelistID: 124,
			},
			{
				RepoID:       1,
				CommitSHA:    commitSHA3,
				ChangelistID: 125,
			},
		}

		got := []repoCommitRow{}

		for rows.Next() {
			var r repoCommitRow
			var hexCommitSHA []byte
			var createdAt time.Time

			if err := rows.Scan(&r.RepoID, &hexCommitSHA, &r.ChangelistID, &createdAt); err != nil {
				t.Fatal(err)
			}

			// All we care is that createdAt has a value, we don't really care about what that is.
			require.NotNil(t, createdAt)

			r.CommitSHA = hex.EncodeToString(hexCommitSHA)
			got = append(got, r)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched rows, diff (-want, +got):\n %v\n", diff)
		}
	})

	t.Run("GetLatestForRepo", func(t *testing.T) {
		t.Run("existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(repoID1))
			require.NoError(t, err, "unexpected error in GetLatestForRepo")
			require.NotNil(t, repoCommit, "repoCommit was not expected to be nil")
			require.Equal(
				t,
				&types.RepoCommit{
					ID:                   3,
					RepoID:               api.RepoID(repoID1),
					CommitSHA:            dbutil.CommitBytea(commitSHA3),
					PerforceChangelistID: 125,
				},
				repoCommit,
				"repoCommit row is not as expected",
			)
		})

		t.Run("non existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(3))
			require.Error(t, err)
			require.True(t, errors.Is(err, sql.ErrNoRows))
			require.Nil(t, repoCommit)
		})
	})

	t.Run("GetRepoCommit", func(t *testing.T) {
		t.Run("existing row", func(t *testing.T) {
			gotRow, err := s.GetRepoCommitChangelist(ctx, 1, 123)
			require.NoError(t, err)
			if diff := cmp.Diff(&types.RepoCommit{
				ID:                   1,
				RepoID:               api.RepoID(1),
				CommitSHA:            dbutil.CommitBytea(commitSHA1),
				PerforceChangelistID: 123,
			}, gotRow); diff != "" {
				t.Errorf("mismatched row, (-want, +got)\n%s", diff)
			}
		})

		t.Run("non existing row", func(t *testing.T) {
			_, err := s.GetRepoCommitChangelist(ctx, 2, 999)
			require.Error(t, err)

			var notFoundError *perforce.ChangelistNotFoundError
			if errors.As(err, &notFoundError) {
				require.Equal(t, api.RepoID(2), notFoundError.RepoID)
				require.Equal(t, int64(999), notFoundError.ID)
			} else {
				t.Fatalf("wrong error type, want ChangelistNotFoundError, got %T", err)
			}
		})
	})

	t.Run("BatchGetRepoCommitChangelist", func(t *testing.T) {
		changelistIds := []RepoChangelistIDs{
			{
				RepoID:        api.RepoID(1),
				ChangelistIDs: []int64{123, 124, 125},
			},
			{
				RepoID:        api.RepoID(2),
				ChangelistIDs: []int64{126},
			},
		}
		t.Run("existing rows", func(t *testing.T) {
			got, err := s.BatchGetRepoCommitChangelist(ctx, changelistIds...)
			require.NoError(t, err)
			// Make sure every items from changelist ids is present in the result.
			for _, ids := range changelistIds {
				for _, id := range ids.ChangelistIDs {
					_, found := got[ids.RepoID][id]
					require.True(t, found, "row for repo %d and changelist %d was not found", ids.RepoID, id)
				}
			}
		})

		t.Run("return empty map if no result", func(t *testing.T) {
			got, err := s.BatchGetRepoCommitChangelist(ctx, RepoChangelistIDs{RepoID: 3})
			require.NoError(t, err)
			require.Len(t, got[3], 0)
		})
	})
}
