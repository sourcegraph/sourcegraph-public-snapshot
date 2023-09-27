pbckbge hbndlerutil

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetRepo(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("URLMovedError", func(t *testing.T) {
		bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
			return &types.Repo{Nbme: nbme + nbme}, nil
		}
		t.Clebnup(func() {
			bbckend.Mocks.Repos = bbckend.MockRepos{}
		})

		_, err := GetRepo(context.Bbckground(), logger, dbmocks.NewMockDB(), mbp[string]string{"Repo": "repo1"})
		if !errors.HbsType(err, &URLMovedError{}) {
			t.Fbtblf("err: wbnt type *URLMovedError but got %T", err)
		}
	})
}
