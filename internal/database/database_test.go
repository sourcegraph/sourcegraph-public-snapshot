package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	useFastPasswordMocks()
}

func TestDBTransactions(t *testing.T) {
	ctx := context.Background()
	t.Run("no transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(sqlDB)

		err := db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
		require.NoError(t, err)

		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("basic transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(sqlDB)

		// Lifetime of tx
		{
			tx, err := db.Repos().Transact(ctx)
			require.NoError(t, err)

			err = tx.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			tx.Done(nil)
		}

		// After committing the transaction, the repo should be visible
		// outisde the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("rolled back transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(sqlDB)

		// Lifetime of tx
		{
			tx, err := db.Repos().Transact(ctx)
			require.NoError(t, err)

			err = tx.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			tx.Done(errors.New("force a rollback"))
		}

		// After rolling back the transaction, the repo should not be visible
		// outside the transaction
		_, err := db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})

	t.Run("nested transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(sqlDB)

		// Lifetime of tx1
		{
			tx1, err := db.Repos().Transact(ctx)
			require.NoError(t, err)

			err = tx1.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx1.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			// Lifetime of tx2
			{
				tx2, err := db.Repos().Transact(ctx)
				require.NoError(t, err)

				err = tx2.Create(ctx, &types.Repo{ID: 2, Name: "test2"})
				require.NoError(t, err)

				// Get inside the transaction should work
				r, err := tx2.Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)

				// Before committing the transaction, repo 2 should not be visible
				// outside the transaction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)
				_, err = tx1.Get(ctx, 2)
				require.Error(t, err)

				tx2.Done(nil)
			}

			// After committing the transaction, repo 2 should be visible
			// outside the transaction
			r, err = db.Repos().Get(ctx, 2)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test2"), r.Name)
			r, err = tx1.Get(ctx, 2)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test2"), r.Name)

			tx1.Done(nil)
		}

		// After committing the transaction, repo 1 should be visible
		// outisde the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("nested transaction rollback works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(sqlDB)

		// Lifetime of tx1
		{
			tx1, err := db.Repos().Transact(ctx)
			require.NoError(t, err)

			err = tx1.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx1.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			// Lifetime of tx2
			{
				tx2, err := db.Repos().Transact(ctx)
				require.NoError(t, err)

				err = tx2.Create(ctx, &types.Repo{ID: 2, Name: "test2"})
				require.NoError(t, err)

				// Get inside the transaction should work
				r, err := tx2.Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)

				// Before committing the transaction, repo 2 should not be visible
				// outside the transaction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)
				_, err = tx1.Get(ctx, 2)
				require.Error(t, err)

				tx2.Done(errors.New("force rollback"))
			}

			// After rolling back the transaction, repo 2 should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)
			_, err = tx1.Get(ctx, 2)
			require.Error(t, err)

			tx1.Done(nil)
		}

		// After committing the transaction, repo 1 should be visible
		// outisde the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})
}
