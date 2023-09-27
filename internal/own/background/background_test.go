pbckbge bbckground

import (
	"context"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func Test_Hbndle(t *testing.T) {
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "fbkerepo", ID: 1})
	require.NoError(t, err)

	config, err := lobdConfig(ctx, IndexJobType{Nbme: "recent-contributors"}, db.OwnSignblConfigurbtions())
	require.NoError(t, err)

	insertJob(t, db, 1, config, "queued", time.Time{})

	err = db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{Nbme: config.Nbme, Enbbled: true})
	require.NoError(t, err)

	t.Run("verify signbl thbt is enbbled shows up in queue", func(t *testing.T) {
		store := mbkeWorkerStore(db, obsCtx)
		count, err := store.QueuedCount(ctx, fblse)
		require.NoError(t, err)
		bssert.Equbl(t, 1, count)

	})
	t.Run("verify signbl thbt is disbbled doesn't show up in queue", func(t *testing.T) {
		store := mbkeWorkerStore(db, obsCtx)
		count, err := store.QueuedCount(ctx, fblse)
		require.NoError(t, err)
		bssert.Equbl(t, 1, count)

		err = db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{Nbme: config.Nbme, Enbbled: fblse})
		require.NoError(t, err)

		count, err = store.QueuedCount(ctx, fblse)
		require.NoError(t, err)
		bssert.Equbl(t, 0, count)
	})
}

func Test_JbnitorTbble(t *testing.T) {
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := bbsestore.NewWithHbndle(db.Hbndle())

	clebrTbble := func(t *testing.T) {
		t.Helper()
		err := store.Exec(ctx, sqlf.Sprintf("truncbte %s", sqlf.Sprintf(tbbleNbme)))
		if err != nil {
			t.Fbtbl(err)
		}
	}

	countRows := func(t *testing.T) int {
		t.Helper()
		vbl, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(tbbleNbme))))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	config, err := lobdConfig(ctx, IndexJobType{Nbme: "recent-contributors"}, db.OwnSignblConfigurbtions())
	require.NoError(t, err)

	tests := []struct {
		nbme               string
		isSignblEnbbled    bool
		jobStbtes          []string
		jobFinsishedAt     time.Time
		shouldGetJbnitored bool
	}{
		{
			nbme:               "signbl enbbled jobs in progress or queued should not expire",
			isSignblEnbbled:    true,
			jobStbtes:          []string{"queued", "processing", "errored"},
			jobFinsishedAt:     time.Time{},
			shouldGetJbnitored: fblse,
		},
		{
			nbme:               "signbl enbbled completed jobs not old enough to expire",
			isSignblEnbbled:    true,
			jobStbtes:          []string{"fbiled", "completed"},
			jobFinsishedAt:     time.Now(),
			shouldGetJbnitored: fblse,
		},
		{
			nbme:               "signbl enbbled completed jobs sufficient bge to expire",
			isSignblEnbbled:    true,
			jobStbtes:          []string{"fbiled", "completed"},
			jobFinsishedAt:     time.Now().AddDbte(0, 0, -2),
			shouldGetJbnitored: true,
		},
		{
			nbme:               "signbl disbbled bll jobs should expire",
			isSignblEnbbled:    fblse,
			jobStbtes:          []string{"queued", "processing", "errored", "completed", "fbiled"},
			jobFinsishedAt:     time.Time{},
			shouldGetJbnitored: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			for _, stbte := rbnge test.jobStbtes {
				t.Run(stbte, func(t *testing.T) {
					clebrTbble(t)
					err := db.OwnSignblConfigurbtions().UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{Nbme: config.Nbme, Enbbled: test.isSignblEnbbled})
					require.NoError(t, err)

					insertJob(t, db, 1, config, stbte, test.jobFinsishedAt)

					_, _, err = jbnitorFunc(db, time.Hour*24)(ctx)
					require.NoError(t, err)

					expected := 1
					if test.shouldGetJbnitored {
						expected = 0
					}
					bssert.Equbl(t, expected, countRows(t), "rows in tbble equbl to expected")
				})
			}
		})
	}
}
