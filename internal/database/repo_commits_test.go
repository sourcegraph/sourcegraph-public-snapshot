package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func TestRepoCommits(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	repos := db.Repos()
	err := repos.Create(ctx, &types.Repo{ID: 1, Name: "foo"})
	require.NoError(t, err, "failed to insert repo")

	repoID := int32(1)

	commitSHA1 := "98d3ec26623660f17f6c298943f55aa339aa894a"
	commitSHA2 := "4b6152a804c4c177f5fe0dfd61e71cacb64d64dd"
	commitSHA3 := "e9c83398bbd4c4e9481fd20f100a7e49d68d7510"

	data := []types.PerforceChangelist{
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

	s := RepoCommitsWith(logger, db)

	err = s.BatchInsertCommitSHAsWithPerforceChangelistID(ctx, api.RepoID(repoID), data)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("BatchInsertCommitSHAsWithPerforceChangelistID", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `SELECT repo_id, commit_sha, perforce_changelist_id FROM repo_commits ORDER by id`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		wantedRows := map[string]int64{
			commitSHA1: 123,
			commitSHA2: 124,
			commitSHA3: 125,
		}

		totalRows := 0
		for rows.Next() {
			var gotRepoID int32
			var gotCommitSHABytea []byte
			var gotChangelistID int64
			if err := rows.Scan(&gotRepoID, &gotCommitSHABytea, &gotChangelistID); err != nil {
				t.Fatal(err)
			}

			gotCommitSHA := hex.EncodeToString(gotCommitSHABytea)

			require.Equal(t, repoID, gotRepoID, "mismatched repoID")
			require.Equal(t, wantedRows[gotCommitSHA], gotChangelistID, "mismatched commitSHA and changelist ID")

			totalRows += 1
		}

		require.Equal(t, 3, totalRows, "mismatched number of rows")
	})

	t.Run("GetLatestForRepo", func(t *testing.T) {
		t.Run("existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(repoID))
			require.NoError(t, err, "unexpected error in GetLatestForRepo")
			require.NotNil(t, repoCommit, "repoCommit was not expected to be nil")
			require.Equal(
				t,
				&types.RepoCommit{
					ID:                   3,
					RepoID:               api.RepoID(repoID),
					CommitSHA:            dbutil.CommitBytea(commitSHA3),
					PerforceChangelistID: 125,
				},
				repoCommit,
				"repoCommit row is not as expected",
			)
		})

		t.Run("non existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(2))
			require.Error(t, err)
			require.True(t, errors.Is(err, sql.ErrNoRows))
			require.Equal(t, &types.RepoCommit{}, repoCommit)
		})
	})
}
