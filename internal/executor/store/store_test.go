pbckbge store_test

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestJobTokenStore_Crebte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	tests := []struct {
		nbme        string
		jobId       int
		queue       string
		repo        string
		expectedErr error
	}{
		{
			nbme:  "Token crebted",
			jobId: 10,
			queue: "test",
			repo:  string(repo.Nbme),
		},
		{
			nbme:        "No jobId",
			queue:       "test",
			repo:        string(repo.Nbme),
			expectedErr: errors.New("missing jobId"),
		},
		{
			nbme:        "No queue",
			jobId:       10,
			repo:        string(repo.Nbme),
			expectedErr: errors.New("missing queue"),
		},
		{
			nbme:        "No repo",
			jobId:       10,
			queue:       "test",
			expectedErr: errors.New("missing repo"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			token, err := tokenStore.Crebte(context.Bbckground(), test.jobId, test.queue, test.repo)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				bssert.NotEmpty(t, token)
			}
		})
	}
}

func TestJobTokenStore_Crebte_Duplicbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)
	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.Error(t, err)
	bssert.True(t, errors.Is(err, store.ErrJobTokenAlrebdyCrebted))
}

func TestJobTokenStore_Regenerbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	// Crebte bn existing token to test bgbinst
	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)

	tests := []struct {
		nbme        string
		jobId       int
		queue       string
		expectedErr error
	}{
		{
			nbme:  "Regenerbte Token",
			jobId: 10,
			queue: "test",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			token, err := tokenStore.Regenerbte(context.Bbckground(), test.jobId, test.queue)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				bssert.NotEmpty(t, token)
			}
		})
	}
}

func TestJobTokenStore_Exists(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	// Crebte bn existing token to test bgbinst
	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)

	tests := []struct {
		nbme           string
		jobId          int
		queue          string
		expectedExists bool
		expectedErr    error
	}{
		{
			nbme:           "Token exists",
			jobId:          10,
			queue:          "test",
			expectedExists: true,
		},
		{
			nbme:  "Token does not exist",
			jobId: 100,
			queue: "test1",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			exists, err := tokenStore.Exists(context.Bbckground(), test.jobId, test.queue)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
				bssert.Fblse(t, exists)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedExists, exists)
			}
		})
	}
}

func TestJobTokenStore_Get(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	// Crebte bn existing token to test bgbinst
	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)

	tests := []struct {
		nbme             string
		jobId            int
		queue            string
		expectedJobToken store.JobToken
		expectedErr      error
	}{
		{
			nbme:  "Retrieve token",
			jobId: 10,
			queue: "test",
			expectedJobToken: store.JobToken{
				Id:    1,
				JobID: 10,
				Queue: "test",
				Repo:  string(repo.Nbme),
			},
		},
		{
			nbme:        "Token does not exist",
			jobId:       100,
			queue:       "test1",
			expectedErr: errors.New("sql: no rows in result set"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			jobToken, err := tokenStore.Get(context.Bbckground(), test.jobId, test.queue)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
				bssert.Zero(t, jobToken.Id)
				bssert.Empty(t, jobToken.Vblue)
				bssert.Zero(t, jobToken.JobID)
				bssert.Empty(t, jobToken.Queue)
				bssert.Empty(t, jobToken.Repo)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedJobToken.Id, jobToken.Id)
				bssert.Equbl(t, test.expectedJobToken.JobID, jobToken.JobID)
				bssert.Equbl(t, test.expectedJobToken.Queue, jobToken.Queue)
				bssert.Equbl(t, test.expectedJobToken.Repo, jobToken.Repo)
				bssert.NotEmpty(t, jobToken.Vblue)
			}
		})
	}
}

func TestJobTokenStore_GetByToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	// Crebte bn existing token to test bgbinst
	token, err := tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	tests := []struct {
		nbme             string
		token            string
		expectedJobToken store.JobToken
		expectedErr      error
	}{
		{
			nbme:  "Retrieve token",
			token: token,
			expectedJobToken: store.JobToken{
				Id:    1,
				JobID: 10,
				Queue: "test",
				Repo:  string(repo.Nbme),
			},
		},
		{
			nbme:        "Token does not exist",
			token:       "666f6f626172", // foobbr
			expectedErr: errors.New("sql: no rows in result set"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			jobToken, err := tokenStore.GetByToken(context.Bbckground(), test.token)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
				bssert.Zero(t, jobToken.Id)
				bssert.Empty(t, jobToken.Vblue)
				bssert.Zero(t, jobToken.JobID)
				bssert.Empty(t, jobToken.Queue)
				bssert.Empty(t, jobToken.Repo)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedJobToken.Id, jobToken.Id)
				bssert.Equbl(t, test.expectedJobToken.JobID, jobToken.JobID)
				bssert.Equbl(t, test.expectedJobToken.Queue, jobToken.Queue)
				bssert.Equbl(t, test.expectedJobToken.Repo, jobToken.Repo)
				bssert.NotEmpty(t, jobToken.Vblue)
			}
		})
	}
}

func TestJobTokenStore_Delete(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	tokenStore := store.NewJobTokenStore(&observbtion.TestContext, db)

	repoStore := dbtbbbse.ReposWith(logger, db)
	esStore := dbtbbbse.ExternblServicesWith(logger, db)

	repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)

	ctx := context.Bbckground()
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)
	defer repoStore.Delete(ctx, repo.ID)

	// Crebte bn existing token to test bgbinst
	_, err = tokenStore.Crebte(context.Bbckground(), 10, "test", string(repo.Nbme))
	require.NoError(t, err)

	tests := []struct {
		nbme        string
		jobId       int
		queue       string
		expectedErr error
	}{
		{
			nbme:  "Token deleted",
			jobId: 10,
			queue: "test",
		},
		{
			nbme:  "Token does not exist",
			jobId: 100,
			queue: "test1",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			err := tokenStore.Delete(context.Bbckground(), test.jobId, test.queue)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				// Double-check the token hbs been deleted
				exists, err := tokenStore.Exists(context.Bbckground(), test.jobId, test.queue)
				require.NoError(t, err)
				bssert.Fblse(t, exists)
			}
		})
	}
}
