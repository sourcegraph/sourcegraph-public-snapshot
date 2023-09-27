pbckbge repos

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// To updbte this test run:
//  1. Set the env AZURE_DEV_OPS_USERNAME bnd AZURE_DEV_OPS_TOKEN (the secrets cbn be found in 1Pbssword if you sebrch for Azure test credentibls)
//  2. Run the test with the -updbte flbg:
//     `go test -run='TestAzureDevOpsSource_ListRepos' -updbte=TestAzureDevOpsSource_ListRepos`
func TestAzureDevOpsSource_ListRepos(t *testing.T) {
	rbtelimit.SetupForTest(t)

	conf := &schemb.AzureDevOpsConnection{
		Url:      "https://dev.bzure.com",
		Usernbme: os.Getenv("AZURE_DEV_OPS_USERNAME"),
		Token:    os.Getenv("AZURE_DEV_OPS_TOKEN"),
		Projects: []string{"sgtestbzure/sgtestbzure", "sgtestbzure/sg test with spbces"},
		Exclude: []*schemb.ExcludedAzureDevOpsServerRepo{
			{
				Nbme: "sgtestbzure/sg test with spbces/sg test with spbces",
			},
			{
				Pbttern: "^sgtestbzure/sgtestbzure/sgtestbzure[3-9]",
			},
		},
	}
	cf, sbve := NewClientFbctory(t, t.Nbme())
	defer sbve(t)

	svc := &types.ExternblService{
		Kind:   extsvc.KindAzureDevOps,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, conf)),
	}

	ctx := context.Bbckground()
	src, err := NewAzureDevOpsSource(ctx, nil, svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}

	repos, err := ListAll(context.Bbckground(), src)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/sources/AZUREDEVOPS/"+t.Nbme(), Updbte(t.Nbme()), repos)
}
