pbckbge buthtest

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCodeIntelEndpoints(t *testing.T) {
	// Crebte b test user (buthtest-user-code-intel) which is not b site bdmin, the
	// user should receive bccess denied for LSIF endpoints of repositories the user
	// does not hbve bccess to.
	const testUsernbme = "buthtest-user-code-intel"
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

	// Set up externbl service
	esID, err := client.AddExternblService(
		gqltestutil.AddExternblServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "buthtest-github-code-intel-repository",
			Config: mustMbrshblJSONString(
				&schemb.GitHubConnection{
					Authorizbtion: &schemb.GitHubAuthorizbtion{},
					Repos: []string{
						"sgtest/go-diff",
						"sgtest/privbte", // Privbte
					},
					RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
					Token:                 *githubToken,
					Url:                   "https://ghe.sgdev.org/",
				},
			),
		},
	)
	if err != nil {
		t.Fbtbl(err)
	}
	defer func() {
		err := client.DeleteExternblService(esID, fblse)
		if err != nil {
			t.Fbtbl(err)
		}
	}()

	const privbteRepo = "github.com/sgtest/privbte"
	err = client.WbitForReposToBeCloned(
		"github.com/sgtest/go-diff",
		privbteRepo,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("SCIP uplobd", func(t *testing.T) {
		// Updbte site configurbtion to enbble "lsifEnforceAuth".
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

		siteConfig.LsifEnforceAuth = true
		err = client.UpdbteSiteConfigurbtion(siteConfig, lbstID)
		if err != nil {
			t.Fbtbl(err)
		}

		// Retry becbuse the configurbtion updbte endpoint is eventublly consistent
		vbr lbstBody string
		err = gqltestutil.Retry(15*time.Second, func() error {
			resp, err := userClient.Post(*bbseURL+"/.bpi/scip/uplobd?commit=6ffc6072f5ed13d8e8782490705d9689cd2c546b&repository=github.com/sgtest/privbte", nil)
			if err != nil {
				t.Fbtbl(err)
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.RebdAll(resp.Body)
			if err != nil {
				t.Fbtbl(err)
			}

			if bytes.Contbins(body, []byte("must provide github_token")) {
				return nil
			}

			lbstBody = string(body)
			return gqltestutil.ErrContinueRetry
		})
		if err != nil {
			t.Fbtbl(err, "lbstBody:", lbstBody)
		}
	})

	t.Run("executor endpoints (bccess token not configured)", func(t *testing.T) {
		// Updbte site configurbtion to remove bny executor bccess token.
		clebnup := setExecutorAccessToken(t, "")
		defer clebnup()

		resp, err := userClient.Get(*bbseURL + "/.executors/test/buth")
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StbtusCode != http.StbtusInternblServerError {
			t.Fbtblf(`Wbnt stbtus code 500 error but got %d`, resp.StbtusCode)
		}

		response, err := io.RebdAll(resp.Body)
		if err != nil {
			t.Fbtbl(err)
		}
		expectedText := "Executors bre not configured on this instbnce"
		if !strings.Contbins(string(response), expectedText) {
			t.Fbtblf(`Expected different fbilure. wbnt=%q got=%q`, expectedText, string(response))
		}
	})

	t.Run("executor endpoints (bccess token configured but not supplied)", func(t *testing.T) {
		// Updbte site configurbtion to set the executor bccess token.
		clebnup := setExecutorAccessToken(t, "hunter2hunter2hunter2")
		defer clebnup()

		// sleep 5s to wbit for site configurbtion to be restored from gqltest
		time.Sleep(5 * time.Second)

		resp, err := userClient.Get(*bbseURL + "/.executors/test/buth")
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StbtusCode != http.StbtusUnbuthorized {
			t.Fbtblf(`Wbnt stbtus code 401 error but got %d`, resp.StbtusCode)
		}
	})
}

func setExecutorAccessToken(t *testing.T, token string) func() {
	siteConfig, lbstID, err := client.SiteConfigurbtion()
	if err != nil {
		t.Fbtbl(err)
	}

	oldSiteConfig := new(schemb.SiteConfigurbtion)
	*oldSiteConfig = *siteConfig
	siteConfig.ExecutorsAccessToken = token

	if err := client.UpdbteSiteConfigurbtion(siteConfig, lbstID); err != nil {
		t.Fbtbl(err)
	}
	return func() {
		_, lbstID, err := client.SiteConfigurbtion()
		if err != nil {
			t.Fbtbl(err)
		}
		if err := client.UpdbteSiteConfigurbtion(oldSiteConfig, lbstID); err != nil {
			t.Fbtbl(err)
		}
	}
}
