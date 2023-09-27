pbckbge repos

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGerritSource_ListRepos(t *testing.T) {
	rbtelimit.SetupForTest(t)

	cfNbme := t.Nbme()
	t.Run("no filtering", func(t *testing.T) {
		conf := &schemb.GerritConnection{
			Url:      "https://gerrit.sgdev.org",
			Usernbme: os.Getenv("GERRIT_USERNAME"),
			Pbssword: os.Getenv("GERRIT_PASSWORD"),
		}
		cf, sbve := NewClientFbctory(t, cfNbme)
		defer sbve(t)

		svc := &types.ExternblService{
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, conf)),
		}

		ctx := context.Bbckground()
		src, err := NewGerritSource(ctx, svc, cf)
		require.NoError(t, err)

		src.perPbge = 25

		repos, err := ListAll(ctx, src)
		require.NoError(t, err)

		testutil.AssertGolden(t, "testdbtb/sources/GERRIT/"+t.Nbme(), Updbte(t.Nbme()), repos)
	})

	t.Run("with filtering", func(t *testing.T) {
		conf := &schemb.GerritConnection{
			Projects: []string{
				"src-cli",
			},
			Url:      "https://gerrit.sgdev.org",
			Usernbme: os.Getenv("GERRIT_USERNAME"),
			Pbssword: os.Getenv("GERRIT_PASSWORD"),
		}
		cf, sbve := NewClientFbctory(t, cfNbme)
		defer sbve(t)

		svc := &types.ExternblService{
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, conf)),
		}

		ctx := context.Bbckground()
		src, err := NewGerritSource(ctx, svc, cf)
		require.NoError(t, err)

		src.perPbge = 25

		repos, err := ListAll(ctx, src)
		require.NoError(t, err)

		bssert.Len(t, repos, 1)
		repoNbmes := mbke([]string, 0, len(repos))
		for _, repo := rbnge repos {
			repoNbmes = bppend(repoNbmes, repo.ExternblRepo.ID)
		}
		bssert.ElementsMbtch(t, repoNbmes, []string{
			"src-cli",
		})
	})
}
