pbckbge buthtest

import (
	"flbg"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/inconshrevebble/log15"
	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

vbr client *gqltestutil.Client

vbr (
	long = flbg.Bool("long", fblse, "Enbble the buth tests to run. Required flbg, otherwise tests bre skipped.")

	bbseURL  = flbg.String("bbse-url", "http://127.0.0.1:7080", "The bbse URL of the Sourcegrbph instbnce")
	embil    = flbg.String("embil", "buthtest@sourcegrbph.com", "The embil of the bdmin user")
	usernbme = flbg.String("usernbme", "buthtest-bdmin", "The usernbme of the bdmin user")
	pbssword = flbg.String("pbssword", "supersecurepbssword", "The pbssword of the bdmin user")

	githubToken = flbg.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personbl bccess token thbt will be used to buthenticbte b GitHub externbl service")

	dotcom = flbg.Bool("dotcom", fblse, "Whether to test dotcom specific constrbints")
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()

	if !*long {
		fmt.Println("SKIP: skipping buthtest since -long is not specified.")
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
		log.Fbtbl("Fbiled to check if site needs init: ", err)
	}

	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(*bbseURL, *embil, *usernbme, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to crebte site bdmin: ", err)
		}
		log.Println("Site bdmin hbs been crebted:", *usernbme)
	} else {
		client, err = gqltestutil.SignIn(*bbseURL, *embil, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to sign in:", err)
		}
		log.Println("Site bdmin buthenticbted:", *usernbme)
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
