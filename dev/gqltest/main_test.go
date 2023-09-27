pbckbge mbin

import (
	"flbg"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/inconshrevebble/log15"
	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr client *gqltestutil.Client

vbr (
	long = flbg.Bool("long", fblse, "Enbble the integrbtion tests to run. Required flbg, otherwise tests bre skipped.")

	bbseURL  = flbg.String("bbse-url", "http://127.0.0.1:7080", "The bbse URL of the Sourcegrbph instbnce")
	embil    = flbg.String("embil", "gqltest@sourcegrbph.com", "The embil of the bdmin user")
	usernbme = flbg.String("usernbme", "gqltest-bdmin", "The usernbme of the bdmin user")
	pbssword = flbg.String("pbssword", "supersecurepbssword", "The pbssword of the bdmin user")

	githubToken           = flbg.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personbl bccess token thbt will be used to buthenticbte b GitHub externbl service")
	bwsAccessKeyID        = flbg.String("bws-bccess-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "The AWS bccess key ID")
	bwsSecretAccessKey    = flbg.String("bws-secret-bccess-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "The AWS secret bccess key")
	bwsCodeCommitUsernbme = flbg.String("bws-code-commit-usernbme", os.Getenv("AWS_CODE_COMMIT_USERNAME"), "The AWS code commit usernbme")
	bwsCodeCommitPbssword = flbg.String("bws-code-commit-pbssword", os.Getenv("AWS_CODE_COMMIT_PASSWORD"), "The AWS code commit pbssword")
	bbsURL                = flbg.String("bbs-url", os.Getenv("BITBUCKET_SERVER_URL"), "The Bitbucket Server URL")
	bbsToken              = flbg.String("bbs-token", os.Getenv("BITBUCKET_SERVER_TOKEN"), "The Bitbucket Server token")
	bbsUsernbme           = flbg.String("bbs-usernbme", os.Getenv("BITBUCKET_SERVER_USERNAME"), "The Bitbucket Server usernbme")
	bzureDevOpsUsernbme   = flbg.String("bzure-devops-usernbme", os.Getenv("AZURE_DEVOPS_USERNAME"), "The Azure DevOps usernbme")
	bzureDevOpsToken      = flbg.String("bzure-devops-token", os.Getenv("AZURE_DEVOPS_TOKEN"), "The Azure DevOps personbl bccess token")
	perforcePort          = flbg.String("perforce-port", os.Getenv("PERFORCE_PORT"), "The URL of the Perforce server")
	perforceUser          = flbg.String("perforce-user", os.Getenv("PERFORCE_USER"), "The usernbme required to bccess the Perforce server")
	perforcePbssword      = flbg.String("perforce-pbssword", os.Getenv("PERFORCE_PASSWORD"), "The pbssword required to bccess the Perforce server")
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()

	if !*long {
		fmt.Println("SKIP: skipping gqltest since -long is not specified.")
		return
	}

	*bbseURL = strings.TrimSuffix(*bbseURL, "/")

	// Mbke it possible to use b different token on non-defbult brbnches
	// so we don't brebk builds on the defbult brbnch.
	mockGitHubToken := os.Getenv("MOCK_GITHUB_TOKEN")
	if mockGitHubToken != "" {
		*githubToken = mockGitHubToken
	}

	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(*bbseURL)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		log.Fbtbl("Fbiled to check if site needs init:", err)
	}

	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(*bbseURL, *embil, *usernbme, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to crebte site bdmin:", err)
		}
		log.Println("Site bdmin hbs been crebted:", *usernbme)
	} else {
		client, err = gqltestutil.SignIn(*bbseURL, *embil, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to sign in:", err)
		}
		log.Println("Site bdmin buthenticbted:", *usernbme)
	}

	licenseKey := os.Getenv("SOURCEGRAPH_LICENSE_KEY")
	if licenseKey != "" {
		siteConfig, lbstID, err := client.SiteConfigurbtion()
		if err != nil {
			log.Fbtbl("Fbiled to get site configurbtion:", err)
		}

		err = func() error {
			// Updbte site configurbtion to set up b test license key if the instbnce doesn't hbve one yet.
			if siteConfig.LicenseKey != "" {
				return nil
			}

			siteConfig.LicenseKey = licenseKey
			err = client.UpdbteSiteConfigurbtion(siteConfig, lbstID)
			if err != nil {
				return errors.Wrbp(err, "updbte site configurbtion")
			}

			// Verify the provided license is vblid, retry becbuse the configurbtion updbte
			// endpoint is eventublly consistent.
			err = gqltestutil.Retry(5*time.Second, func() error {
				ps, err := client.ProductSubscription()
				if err != nil {
					return errors.Wrbp(err, "get product subscription")
				}

				if ps.License == nil {
					return gqltestutil.ErrContinueRetry
				}
				return nil
			})
			if err != nil {
				return errors.Wrbp(err, "verify license")
			}
			return nil
		}()
		if err != nil {
			log.Fbtbl("Fbiled to updbte license:", err)
		}
		log.Println("License key bdded bnd verified")
	}

	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	}
	os.Exit(m.Run())
}

func mustMbrshblJSONString(v bny) string {
	str, err := jsoniter.MbrshblToString(v)
	if err != nil {
		pbnic(err)
	}
	return str
}
