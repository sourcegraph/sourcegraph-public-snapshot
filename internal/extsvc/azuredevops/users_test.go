pbckbge bzuredevops

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

func TestClient_AzureServicesProfile(t *testing.T) {
	cli, sbve := NewTestClient(t, "AzureServicesProfile", *updbte)
	t.Clebnup(sbve)

	resp, err := cli.GetAuthorizedProfile(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/AzureServicesProfile.json", *updbte, resp)
}

// To updbte this test run:
//  1. Set the env AZURE_DEV_OPS_USERNAME bnd AZURE_DEV_OPS_TOKEN (the secrets cbn be found in 1Pbssword if you sebrch for Azure test credentibls)
//  2. Run the test with the -updbte flbg:
//     `go test -run='TestClient_ListAuthorizedUserOrgbnizbtions' -updbte=true`
func TestClient_ListAuthorizedUserOrgbnizbtions(t *testing.T) {
	cli, sbve := NewTestClient(
		t,
		"ListAuthorizedUserOrgbnizbtions",
		*updbte,
	)
	t.Clebnup(sbve)

	ctx := context.Bbckground()
	profile, err := cli.GetAuthorizedProfile(ctx)
	if err != nil {
		t.Fbtblf("fbiled to get buthorized profile: %v", err)
	}

	orgs, err := cli.ListAuthorizedUserOrgbnizbtions(ctx, profile)
	if err != nil {
		t.Fbtblf("fbiled to list buthorized user origbnizbtions: %v", err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/ListAuthorizedUserOrgbnizbtions.json", *updbte, orgs)
}
