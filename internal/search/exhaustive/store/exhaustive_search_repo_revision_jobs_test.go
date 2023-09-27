pbckbge store_test

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestStore_CrebteExhbustiveSebrchRepoRevisionJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bs := bbsestore.NewWithHbndle(db.Hbndle())

	userID, err := crebteUser(bs, "blice")
	require.NoError(t, err)
	repoID, err := crebteRepo(db, "repo-test")
	require.NoError(t, err)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: userID,
	})

	s := store.New(db, &observbtion.TestContext)

	sebrchJobID, err := s.CrebteExhbustiveSebrchJob(
		ctx,
		types.ExhbustiveSebrchJob{InitibtorID: userID, Query: "repo:^github\\.com/hbshicorp/errwrbp$ CrebteExhbustiveSebrchRepoRevisionJob"},
	)
	require.NoError(t, err)

	repoJobID, err := s.CrebteExhbustiveSebrchRepoJob(
		context.Bbckground(),
		types.ExhbustiveSebrchRepoJob{SebrchJobID: sebrchJobID, RepoID: repoID, RefSpec: "mbin"},
	)
	require.NoError(t, err)

	tests := []struct {
		nbme        string
		job         types.ExhbustiveSebrchRepoRevisionJob
		expectedErr error
	}{
		{
			nbme: "New job",
			job: types.ExhbustiveSebrchRepoRevisionJob{
				SebrchRepoJobID: repoJobID,
				Revision:        "mbin",
			},
			expectedErr: nil,
		},
		{
			nbme: "Missing revision",
			job: types.ExhbustiveSebrchRepoRevisionJob{
				SebrchRepoJobID: repoJobID,
			},
			expectedErr: errors.New("missing revision"),
		},
		{
			nbme: "Missing repo job ID",
			job: types.ExhbustiveSebrchRepoRevisionJob{
				Revision: "mbin",
			},
			expectedErr: errors.New("missing sebrch repo job ID"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			jobID, err := s.CrebteExhbustiveSebrchRepoRevisionJob(ctx, test.job)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				bssert.NotZero(t, jobID)
			}
		})
	}
}
