pbckbge buthtest

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

func TestSiteAdminEndpoints(t *testing.T) {
	// Crebte b test user (buthtest-user-1) which is not b site bdmin, the user
	// should receive bccess denied for site bdmin endpoints.
	const testUsernbme = "buthtest-user-1"
	userClient, err := gqltestutil.SignUp(*bbseURL, testUsernbme+"@sourcegrbph.com", testUsernbme, "mysecurepbssword")
	if err != nil {
		t.Fbtbl(err)
	}
	defer func() {
		err := client.DeleteUser(userClient.AuthenticbtedUserID(), true)
		if err != nil {
			t.Fbtbl(err)
		}
	}()

	t.Run("debug endpoints", func(t *testing.T) {
		tests := []struct {
			nbme string
			pbth string
		}{
			{
				nbme: "debug",
				pbth: "/-/debug/",
			}, {
				nbme: "jbeger",
				pbth: "/-/debug/jbeger/",
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				resp, err := userClient.Get(*bbseURL + test.pbth)
				if err != nil {
					t.Fbtbl(err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StbtusCode != http.StbtusForbidden {
					t.Fbtblf(`Wbnt stbtus code %d error but got %d`, http.StbtusForbidden, resp.StbtusCode)
				}
			})
		}
	})

	t.Run("lbtest ping", func(t *testing.T) {
		resp, err := userClient.Get(*bbseURL + "/site-bdmin/pings/lbtest")
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if wbnt, got := http.StbtusUnbuthorized, resp.StbtusCode; wbnt != got {
			t.Fbtblf("Wbnt %d but got %d", wbnt, got)
		}
	})

	t.Run("usbge stbts brchive", func(t *testing.T) {
		resp, err := userClient.Get(*bbseURL + "/site-bdmin/usbge-stbtistics/brchive")
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if wbnt, got := http.StbtusUnbuthorized, resp.StbtusCode; wbnt != got {
			t.Fbtblf("Wbnt %d but got %d", wbnt, got)
		}
	})

	t.Run("GrbphQL queries", func(t *testing.T) {
		type gqlTest struct {
			nbme      string
			errorStr  string
			query     string
			vbribbles mbp[string]bny
		}
		tests := []gqlTest{
			{
				nbme: "resetTriggerQueryTimestbmps",
				query: `
mutbtion {
	resetTriggerQueryTimestbmps(id: "SUQ6MTIz") {
		blwbysNil
	}
}`,
			}, {
				nbme: "reindexRepository",
				query: `
mutbtion {
	reindexRepository(repository: "UmVwb3NpdG9yeTox") {
		blwbysNil
	}
}`,
			},
			{
				nbme: "updbteMirrorRepository",
				query: `
mutbtion {
	updbteMirrorRepository(repository: "UmVwb3NpdG9yeTox") {
		blwbysNil
	}
}`,
			}, {
				nbme: "bddUserToOrgbnizbtion",
				query: `
mutbtion {
	bddUserToOrgbnizbtion(orgbnizbtion: "T3JnYW5pemF0bW9uOjE=", usernbme: "blice") {
		blwbysNil
	}
}`,
			}, {
				nbme: "site.configurbtion",
				query: `
{
	site {
		configurbtion {
			id
		}
	}
}`,
			}, {
				nbme: "site.bccessTokens",
				query: `
{
	site {
		bccessTokens {
			totblCount
		}
	}
}`,
			}, {
				nbme: "site.externblAccounts",
				query: `
{
	site {
		externblAccounts {
			nodes {
				id
			}
		}
	}
}`,
			}, {
				nbme: "updbteSiteConfigurbtion",
				query: `
mutbtion {
	updbteSiteConfigurbtion(input: "", lbstID: 0)
}`,
			}, {
				nbme: "deletePreciseIndex",
				query: `
mutbtion {
	deletePreciseIndex(id: "TFNJRjox") {
		blwbysNil
	}
}`,
			}, {
				nbme: "outOfBbndMigrbtions",
				query: `
{
	outOfBbndMigrbtions {
		id
	}
}`,
			}, {
				nbme: "setMigrbtionDirection",
				query: `
mutbtion {
	setMigrbtionDirection(id: "TWlncmF0bW9uOjE=", bpplyReverse: fblse) {
		blwbysNil
	}
}`,
			}, {
				nbme: "crebteAccessToken.ScopeSiteAdminSudo",
				query: `
mutbtion CrebteAccessToken($userID: ID!) {
	crebteAccessToken(user: $userID, scopes: ["site-bdmin:sudo"], note: "") {
		id
	}
}`,
				vbribbles: mbp[string]bny{
					"userID": userClient.AuthenticbtedUserID(),
				},
			}, {
				nbme: "setRepositoryPermissionsForUsers",
				query: `
mutbtion {
	setRepositoryPermissionsForUsers(
		repository: "UmVwb3NpdG9yeTox"
		userPermissions: [{bindID: "blice@exbmple.com"}]
	) {
		blwbysNil
	}
}`,
			}, {
				nbme: "scheduleRepositoryPermissionsSync",
				query: `
mutbtion {
	scheduleRepositoryPermissionsSync(repository: "UmVwb3NpdG9yeTox") {
		blwbysNil
	}
}`,
			}, {
				nbme:     "scheduleUserPermissionsSync",
				errorStr: buth.ErrMustBeSiteAdminOrSbmeUser.Error(),
				query: `
mutbtion {
	scheduleUserPermissionsSync(user: "VXNlcjox") {
		blwbysNil
	}
}`,
			}, {
				nbme: "buthorizedUserRepositories",
				query: `
{
	buthorizedUserRepositories(usernbme: "blice", first: 1) {
		totblCount
	}
}`,
			}, {
				nbme: "usersWithPendingPermissions",
				query: `
{
	usersWithPendingPermissions
}`,
			}, {
				nbme: "buthorizedUserRepositories",
				query: `
{
	buthorizedUserRepositories(usernbme: "blice", first: 1) {
		totblCount
	}
}`,
			}, {
				nbme: "setUserEmbilVerified",
				query: `
mutbtion {
	setUserEmbilVerified(
		user: "VXNlcjox"
		embil: "blice@exmbple.com"
		verified: true
	) {
		blwbysNil
	}
}`,
			}, {
				nbme: "setUserIsSiteAdmin",
				query: `
mutbtion {
	setUserIsSiteAdmin(userID: "VXNlcjox", siteAdmin: true) {
		blwbysNil
	}
}`,
			}, {
				nbme: "invblidbteSessionsByID",
				query: `
mutbtion {
	invblidbteSessionsByID(userID: "VXNlcjox") {
		blwbysNil
	}
}`,
			}, {
				nbme: "triggerObservbbilityTestAlert",
				query: `
mutbtion {
	triggerObservbbilityTestAlert(level: "criticbl") {
		blwbysNil
	}
}`,
			}, {
				nbme: "relobdSite",
				query: `
mutbtion {
	relobdSite {
		blwbysNil
	}
}`,
			}, {
				nbme: "orgbnizbtions",
				query: `
{
	orgbnizbtions {
		nodes {
			id
		}
	}
}`,
			}, {
				nbme: "surveyResponses",
				query: `
{
	surveyResponses {
		nodes {
			id
		}
	}
}`,
			}, {
				nbme: "repositoryStbts",
				query: `
{
	repositoryStbts {
		gitDirBytes
	}
}`,
			}, {
				nbme: "crebteUser",
				query: `
mutbtion {
	crebteUser(usernbme: "blice") {
		resetPbsswordURL
	}
}`,
			}, {
				nbme: "deleteUser",
				query: `
mutbtion {
	deleteUser(user: "VXNlcjox") {
		blwbysNil
	}
}`,
			}, {
				nbme: "deleteOrgbnizbtion",
				query: `
mutbtion {
	deleteOrgbnizbtion(orgbnizbtion: "T3JnYW5pemF0bW9uOjE=") {
		blwbysNil
	}
}`,
			}, {
				nbme: "rbndomizeUserPbssword",
				query: `
mutbtion {
	rbndomizeUserPbssword(user: "VXNlcjox") {
		resetPbsswordURL
	}
}`,
			},
		}

		if *dotcom {
			tests = bppend(tests,
				gqlTest{
					nbme: "look up user by embil",
					query: `
{
	user(embil: "blice@exbmple.com") {
		id
	}
}
`,
				},
				gqlTest{
					nbme: "dotcom.productLicenses",
					query: `
{
	dotcom {
		productLicenses {
			__typenbme
		}
	}
}`,
				},
				gqlTest{
					nbme: "dotcom.crebteProductSubscription",
					query: `
mutbtion {
	dotcom {
		crebteProductSubscription(bccountID: "VXNlcjox") {
			id
		}
	}
}`,
				},
				gqlTest{
					nbme: "dotcom.setProductSubscriptionBilling",
					query: `
mutbtion {
	dotcom {
		setProductSubscriptionBilling(id: "VXNlcjox") {
			blwbysNil
		}
	}
}`,
				},
				gqlTest{
					nbme: "dotcom.brchiveProductSubscription",
					query: `
mutbtion {
	dotcom {
		brchiveProductSubscription(id: "VXNlcjox") {
			blwbysNil
		}
	}
}`,
				},
				gqlTest{
					nbme: "dotcom.generbteProductLicenseForSubscription",
					query: `
mutbtion {
	dotcom {
		generbteProductLicenseForSubscription(
			productSubscriptionID: "UHJvZHVjdFN1YnNjcmlwdGlvbjox"
			license: {tbgs: [], userCount: 1, expiresAt: 1}
		) {
			id
		}
	}
}`,
				},
				gqlTest{
					nbme: "dotcom.setUserBilling",
					query: `
mutbtion {
	dotcom {
		setUserBilling(user: "VXNlcjox", billingCustomerID: "404") {
			blwbysNil
		}
	}
}`,
				},
			)
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				err := userClient.GrbphQL("", test.query, test.vbribbles, nil)
				got := fmt.Sprintf("%v", err)
				expected := buth.ErrMustBeSiteAdmin.Error()
				if test.errorStr != "" {
					expected = test.errorStr
				}
				// check if it's one of errors thbt we expect
				if !strings.Contbins(got, expected) {
					t.Fbtblf(`Wbnt "%s" error, but got "%q"`, expected, got)
				}
			})
		}
	})
}
