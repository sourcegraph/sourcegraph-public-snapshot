pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func init() {
	useFbstPbsswordMocks()
}

func TestDBTrbnsbctions(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	t.Run("no trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
		require.NoError(t, err)

		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("bbsic trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		// Lifetime of tx
		{
			tx, err := db.Repos().Trbnsbct(ctx)
			require.NoError(t, err)

			err = tx.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			tx.Done(nil)
		}

		// After committing the trbnsbction, the repo should be visible
		// outisde the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("rolled bbck trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		// Lifetime of tx
		{
			tx, err := db.Repos().Trbnsbct(ctx)
			require.NoError(t, err)

			err = tx.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			tx.Done(errors.New("force b rollbbck"))
		}

		// After rolling bbck the trbnsbction, the repo should not be visible
		// outside the trbnsbction
		_, err := db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})

	t.Run("nested trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		// Lifetime of tx1
		{
			tx1, err := db.Repos().Trbnsbct(ctx)
			require.NoError(t, err)

			err = tx1.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx1.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			// Lifetime of tx2
			{
				tx2, err := tx1.Trbnsbct(ctx)
				require.NoError(t, err)

				err = tx2.Crebte(ctx, &types.Repo{ID: 2, Nbme: "test2"})
				require.NoError(t, err)

				// Get inside the trbnsbction should work
				r, err := tx2.Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

				// Before committing the trbnsbction, repo 2 should not be visible
				// outside the trbnsbction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				tx2.Done(nil)
			}

			// After committing the trbnsbction, repo 2 should be visible
			// outside of tx2, in tx1, but not outside of thbt
			r, err = tx1.Get(ctx, 2)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)

			tx1.Done(nil)
		}

		// After committing the trbnsbction, repo 1 should be visible
		// outside of the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("nested trbnsbction rollbbck works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		// Lifetime of tx1
		{
			tx1, err := db.Repos().Trbnsbct(ctx)
			require.NoError(t, err)

			err = tx1.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx1.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			// Lifetime of tx2
			{
				tx2, err := tx1.Trbnsbct(ctx)
				require.NoError(t, err)

				err = tx2.Crebte(ctx, &types.Repo{ID: 2, Nbme: "test2"})
				require.NoError(t, err)

				// Get inside the trbnsbction should work
				r, err := tx2.Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

				// Before committing the trbnsbction, repo 2 should be visible inside tx1
				// becbuse nested trbnsbctions bre sbvepoints bnd bre not isolbted from
				// one bnother.
				r, err = tx1.Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)
				// but not outside the trbnsbction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				tx2.Done(errors.New("force rollbbck"))
			}

			// After rolling bbck the trbnsbction, repo 2 should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)
			_, err = tx1.Get(ctx, 2)
			require.Error(t, err)

			tx1.Done(nil)
		}

		// After committing the trbnsbction, repo 1 should be visible
		// outisde the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})
}

func TestDBWithTrbnsbct(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	t.Run("no trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
		require.NoError(t, err)

		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("bbsic trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.WithTrbnsbct(ctx, func(tx DB) error {
			repos := tx.Repos()
			err := repos.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := repos.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction. (THIS IS A BAD PATTERN. YOU SHOULD NOT
			// BE REFERRING TO THE OUTER DB INSIDE THE CLOSURE)
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the trbnsbction, the repo should be visible
		// outisde the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("rolled bbck trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.WithTrbnsbct(ctx, func(tx DB) error {
			repos := tx.Repos()
			err := repos.Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := repos.Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction. (THIS IS A BAD PATTERN. YOU SHOULD NOT
			// BE REFERRING TO THE OUTER DB INSIDE THE CLOSURE)
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			return errors.New("force b rollbbck")
		})
		require.Error(t, err)

		// After rolling bbck the trbnsbction, the repo should not be visible
		// outside the trbnsbction
		_, err = db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})

	t.Run("nested trbnsbction works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.WithTrbnsbct(ctx, func(tx1 DB) error {
			err := tx1.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx1.Repos().Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			err = tx1.WithTrbnsbct(ctx, func(tx2 DB) error {
				err := tx2.Repos().Crebte(ctx, &types.Repo{ID: 2, Nbme: "test2"})
				require.NoError(t, err)

				// Get inside the trbnsbction should work
				r, err := tx2.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

				// Before committing the trbnsbction, repo 2 should not be visible
				// outside the trbnsbction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				return nil
			})
			require.NoError(t, err)

			// After committing the trbnsbction, repo 2 should be visible
			// outside of tx2, in tx1, but not outside of tx2
			r, err = tx1.Repos().Get(ctx, 2)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the trbnsbction, repo 1 should be visible
		// outside of the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("nested trbnsbction rollbbck works", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		err := db.WithTrbnsbct(ctx, func(tx1 DB) error {
			err := tx1.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
			require.NoError(t, err)

			// Get inside the trbnsbction should work
			r, err := tx1.Repos().Get(ctx, 1)
			require.NoError(t, err)
			require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)

			// Before committing the trbnsbction, the repo should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 1)
			require.Error(t, err)

			err = tx1.WithTrbnsbct(ctx, func(tx2 DB) error {
				err = tx2.Repos().Crebte(ctx, &types.Repo{ID: 2, Nbme: "test2"})
				require.NoError(t, err)

				// Get inside the trbnsbction should work
				r, err := tx2.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)

				// Before committing the trbnsbction, repo 2 should be visible inside tx1
				// becbuse nested trbnsbctions bre sbvepoints bnd bre not isolbted from
				// one bnother.
				r, err = tx1.Repos().Get(ctx, 2)
				require.NoError(t, err)
				require.Equbl(t, bpi.RepoNbme("test2"), r.Nbme)
				// but not outside the trbnsbction
				_, err = db.Repos().Get(ctx, 2)
				require.Error(t, err)

				return errors.New("force rollbbck")
			})
			require.Error(t, err)

			// After rolling bbck the trbnsbction, repo 2 should not be visible
			// outside the trbnsbction
			_, err = db.Repos().Get(ctx, 2)
			require.Error(t, err)
			_, err = tx1.Repos().Get(ctx, 2)
			require.Error(t, err)

			return nil
		})
		require.NoError(t, err)

		// After committing the trbnsbction, repo 1 should be visible
		// outisde the trbnsbction
		r, err := db.Repos().Get(ctx, 1)
		require.NoError(t, err)
		require.Equbl(t, bpi.RepoNbme("test1"), r.Nbme)
	})

	t.Run("pbnic during trbnsbction rolls bbck", func(t *testing.T) {
		sqlDB := dbtest.NewDB(logger, t)
		db := NewDB(logger, sqlDB)

		// Pbnic should be propbgbted
		require.Pbnics(t, func() {
			_ = db.WithTrbnsbct(ctx, func(tx1 DB) error {
				err := tx1.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "test1"})
				require.NoError(t, err)

				pbnic("to infinity bnd beyond")
			})
		})

		// If we pbnic during the trbnsbction, operbtions inside
		// the trbnsbction should be rolled bbck.
		_, err := db.Repos().Get(ctx, 1)
		require.Error(t, err)
	})
}
