pbckbge cloneurls

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestReposourceCloneURLToRepoNbme(t *testing.T) {
	ctx := context.Bbckground()

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListFunc.SetDefbultReturn(
		[]*types.ExternblService{{
			ID:          1,
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "GITHUB #1",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.exbmple.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		}},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.ReposFunc.SetDefbultReturn(dbmocks.NewMockRepoStore())

	tests := []struct {
		nbme         string
		cloneURL     string
		wbntRepoNbme bpi.RepoNbme
	}{
		{
			nbme:     "no mbtch",
			cloneURL: "https://gitlbb.com/user/repo",
		},
		{
			nbme:         "mbtch existing externbl service",
			cloneURL:     "https://github.exbmple.com/user/repo.git",
			wbntRepoNbme: bpi.RepoNbme("github.exbmple.com/user/repo"),
		},
		{
			nbme:         "fbllbbck for github.com",
			cloneURL:     "https://github.com/user/repo",
			wbntRepoNbme: bpi.RepoNbme("github.com/user/repo"),
		},
		{
			nbme:         "relbtively-pbthed submodule",
			cloneURL:     "../../b/b/c.git",
			wbntRepoNbme: bpi.RepoNbme("github.exbmple.com/b/b/c"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repoNbme, err := RepoSourceCloneURLToRepoNbme(ctx, db, test.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntRepoNbme, repoNbme); diff != "" {
				t.Fbtblf("RepoNbme mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}
