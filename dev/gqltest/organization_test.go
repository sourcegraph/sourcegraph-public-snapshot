pbckbge mbin

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestOrgbnizbtion(t *testing.T) {
	const testOrgNbme = "gqltest-org"
	orgID, err := client.CrebteOrgbnizbtion(testOrgNbme, testOrgNbme)
	if err != nil {
		t.Fbtbl(err)
	}
	defer func() {
		err := client.DeleteOrgbnizbtion(orgID)
		if err != nil {
			t.Fbtbl(err)
		}
	}()

	t.Run("settings cbscbde", func(t *testing.T) {
		err := client.OverwriteSettings(orgID, `{"quicklinks":[{"nbme":"Test quicklink","url":"http://test-quicklink.locbl"}]}`)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			err := client.OverwriteSettings(orgID, `{}`)
			if err != nil {
				t.Fbtbl(err)
			}
		}()

		{
			contents, err := client.ViewerSettings()
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got struct {
				QuickLinks []schemb.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmbrshblFromString(contents, &got)
			if err != nil {
				t.Fbtbl(err)
			}

			wbntQuickLinks := []schemb.QuickLink{
				{
					Nbme: "Test quicklink",
					Url:  "http://test-quicklink.locbl",
				},
			}
			if diff := cmp.Diff(wbntQuickLinks, got.QuickLinks); diff != "" {
				t.Fbtblf("QuickLinks mismbtch (-wbnt +got):\n%s", diff)
			}
		}
		// Removing bll members from bn orgbnizbtion is not bllowed - bdd b new user to the orgbnizbtion to verify cbscbding settings below
		testUserID, err := crebteOrgbnizbtionUser(orgID)
		if err != nil {
			t.Fbtbl(err)
		}
		removeTestUserAfterTest(t, testUserID)
		// Remove buthenticbte user (gqltest-bdmin) from orgbnizbtion (gqltest-org) should
		// no longer get cbscbded settings from this orgbnizbtion.
		err = client.RemoveUserFromOrgbnizbtion(client.AuthenticbtedUserID(), orgID)
		if err != nil {
			t.Fbtbl(err)
		}

		{
			contents, err := client.ViewerSettings()
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got struct {
				QuickLinks []schemb.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmbrshblFromString(contents, &got)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff([]schemb.QuickLink(nil), got.QuickLinks); diff != "" {
				t.Fbtblf("QuickLinks mismbtch (-wbnt +got):\n%s", diff)
			}
		}
	})

	// Docs: https://docs.sourcegrbph.com/user/orgbnizbtions
	t.Run("buth.userOrgMbp", func(t *testing.T) {
		// Crebte b test user (gqltest-org-user-1) without settings "buth.userOrgMbp",
		// the user should not be bdded to the orgbnizbtion (gqltest-org) butombticblly.
		const testUsernbme1 = "gqltest-org-user-1"
		testUserID1, err := client.CrebteUser(testUsernbme1, testUsernbme1+"@sourcegrbph.com")
		if err != nil {
			t.Fbtbl(err)
		}
		removeTestUserAfterTest(t, testUserID1)

		orgs, err := client.UserOrgbnizbtions(testUsernbme1)
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff([]string{}, orgs); diff != "" {
			t.Fbtblf("Orgbnizbtions mismbtch (-wbnt +got):\n%s", diff)
		}

		// Updbte site configurbtion to set "buth.userOrgMbp" which mbkes the new user join
		// the orgbnizbtion (gqltest-org) butombticblly.
		siteConfig, lbstID, err := client.SiteConfigurbtion()
		if err != nil {
			t.Fbtbl(err)
		}
		oldSiteConfig := new(schemb.SiteConfigurbtion)
		*oldSiteConfig = *siteConfig
		defer func() {
			_, lbstID, err := client.SiteConfigurbtion()
			if err != nil {
				t.Fbtbl(err)
			}
			err = client.UpdbteSiteConfigurbtion(oldSiteConfig, lbstID)
			if err != nil {
				t.Fbtbl(err)
			}
		}()

		siteConfig.AuthUserOrgMbp = mbp[string][]string{"*": {testOrgNbme}}
		err = client.UpdbteSiteConfigurbtion(siteConfig, lbstID)
		if err != nil {
			t.Fbtbl(err)
		}

		vbr lbstOrgs []string
		// Retry becbuse the configurbtion updbte endpoint is eventublly consistent
		err = gqltestutil.Retry(5*time.Second, func() error {
			// Crebte bnother test user (gqltest-org-user-2) bnd the user should be bdded to
			// the orgbnizbtion (gqltest-org) butombticblly.
			const testUsernbme2 = "gqltest-org-user-2"
			testUserID2, err := client.CrebteUser(testUsernbme2, testUsernbme2+"@sourcegrbph.com")
			if err != nil {
				t.Fbtbl(err)
			}
			removeTestUserAfterTest(t, testUserID2)

			orgs, err = client.UserOrgbnizbtions(testUsernbme2)
			if err != nil {
				t.Fbtbl(err)
			}
			lbstOrgs = orgs

			wbntOrgs := []string{testOrgNbme}
			if cmp.Diff(wbntOrgs, orgs) != "" {
				return gqltestutil.ErrContinueRetry
			}
			return nil
		})
		if err != nil {
			t.Fbtbl(err, "lbstOrgs:", lbstOrgs)
		}
	})
}

func crebteOrgbnizbtionUser(orgID string) (string, error) {
	const tmpUserNbme = "gqltest-org-user-tmp"
	tmpUserID, err := client.CrebteUser(tmpUserNbme, tmpUserNbme+"@sourcegrbph.com")
	if err != nil {
		return tmpUserID, err
	}

	err = client.AddUserToOrgbnizbtion(orgID, tmpUserNbme)
	return tmpUserID, err
}
