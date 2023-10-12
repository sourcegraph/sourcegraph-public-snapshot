package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

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
	logger := logtest.Scoped(t)
	t.Run("no transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
		require.NoError(t, err)

		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("basic transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

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
		db := NewDB(logger, sqlDB)

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
		db := NewDB(logger, sqlDB)

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
				tx2, err := tx1.Transact(ctx)
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

				tx2.Done(nil)
			}

			// After committing the transaction, repo 2 should be visible
			// outside of tx2, in tx1, but not outside of that
			r, err = tx1.Get(ctx, 2)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test2"), r.Name)

			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)

			tx1.Done(nil)
		}

		// After committing the transaction, repo 1 should be visible
		// outside of the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("nested transaction rollback works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

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
				tx2, err := tx1.Transact(ctx)
				require.NoError(t, err)

				err = tx2.Create(ctx, &types.Repo{ID: 2, Name: "test2"})
				require.NoError(t, err)

				// Get inside the transaction should work
				r, err := tx2.Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)

				// Before committing the transaction, repo 2 should be visible inside tx1
				// because nested transactions are savepoints and are not isolated from
				// one another.
				r, err = tx1.Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)
				// but not outside the transaction
				_, err = db.Repos().Get(ctx, 2)
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

func TestDBWithTransact(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	t.Run("no transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
		require.NoError(t, err)

		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("basic transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.WithTransact(ctx, func(tx DB) error {
			repos := tx.Repos()
			err := repos.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := repos.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction. (THIS IS A BAD PATTERN. YOU SHOULD NOT
			// BE REFERRING TO THE OUTER DB INSIDE THE CLOSURE)
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the transaction, the repo should be visible
		// outisde the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("rolled back transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.WithTransact(ctx, func(tx DB) error {
			repos := tx.Repos()
			err := repos.Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := repos.Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction. (THIS IS A BAD PATTERN. YOU SHOULD NOT
			// BE REFERRING TO THE OUTER DB INSIDE THE CLOSURE)
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			return errors.New("force a rollback")
		})
		require.Error(t, err)

		// After rolling back the transaction, the repo should not be visible
		// outside the transaction
		_, err = db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})

	t.Run("nested transaction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.WithTransact(ctx, func(tx1 DB) error {
			err := tx1.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx1.Repos().Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			err = tx1.WithTransact(ctx, func(tx2 DB) error {
				err := tx2.Repos().Create(ctx, &types.Repo{ID: 2, Name: "test2"})
				require.NoError(t, err)

				// Get inside the transaction should work
				r, err := tx2.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)

				// Before committing the transaction, repo 2 should not be visible
				// outside the transaction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				return nil
			})
			require.NoError(t, err)

			// After committing the transaction, repo 2 should be visible
			// outside of tx2, in tx1, but not outside of tx2
			r, err = tx1.Repos().Get(ctx, 2)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test2"), r.Name)

			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the transaction, repo 1 should be visible
		// outside of the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("nested transaction rollback works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		err := db.WithTransact(ctx, func(tx1 DB) error {
			err := tx1.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
			require.NoError(t, err)

			// Get inside the transaction should work
			r, err := tx1.Repos().Get(ctx, 1)
			require.NoError(t, err)
			require.Equal(t, api.RepoName("test1"), r.Name)

			// Before committing the transaction, the repo should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			err = tx1.WithTransact(ctx, func(tx2 DB) error {
				err = tx2.Repos().Create(ctx, &types.Repo{ID: 2, Name: "test2"})
				require.NoError(t, err)

				// Get inside the transaction should work
				r, err := tx2.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)

				// Before committing the transaction, repo 2 should be visible inside tx1
				// because nested transactions are savepoints and are not isolated from
				// one another.
				r, err = tx1.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equal(t, api.RepoName("test2"), r.Name)
				// but not outside the transaction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				return errors.New("force rollback")
			})
			require.Error(t, err)

			// After rolling back the transaction, repo 2 should not be visible
			// outside the transaction
			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)
			_, err = tx1.Repos().Get(ctx, 2)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the transaction, repo 1 should be visible
		// outisde the transaction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, api.RepoName("test1"), r.Name)
	})

	t.Run("panic during transaction rolls back", func(t *testing.T) {
		sqlDB := dbtest.NewDB(t)
		db := NewDB(logger, sqlDB)

		// Panic should be propagated
		require.Panics(t, func() {
			_ = db.WithTransact(ctx, func(tx1 DB) error {
				err := tx1.Repos().Create(ctx, &types.Repo{ID: 1, Name: "test1"})
				require.NoError(t, err)

				panic("to infinity and beyond")
			})
		})

		// If we panic during the transaction, operations inside
		// the transaction should be rolled back.
		_, err := db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})
}
