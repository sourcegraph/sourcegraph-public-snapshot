pbckbge mbin

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

type computeClient interfbce {
	Compute(query string) ([]gqltestutil.MbtchContext, error)
}

func testComputeClient(t *testing.T, client computeClient) {
	t.Run("compute endpoint returns results", func(t *testing.T) {
		results, err := client.Compute(`repo:^github.com/sgtest/go-diff$ file:\.go func Pbrse(\w+)`)
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
		if len(results) == 0 {
			t.Error("Expected results, got none")
		}

	})
}

func TestCompute(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment vbribble GITHUB_TOKEN is not set")
	}

	// Set up externbl service
	_, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "gqltest-github-sebrch",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/jbvb-lbngserver",
				"sgtest/jsonrpc2",
				"sgtest/go-diff",
				"sgtest/bppdbsh",
				"sgtest/sourcegrbph-typescript",
			},
			RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
		}),
	})
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.WbitForReposToBeCloned(
		"github.com/sgtest/jbvb-lbngserver",
		"github.com/sgtest/jsonrpc2",
		"github.com/sgtest/go-diff",
		"github.com/sgtest/bppdbsh",
		"github.com/sgtest/sourcegrbph-typescript",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.WbitForReposToBeIndexed(
		"github.com/sgtest/jbvb-lbngserver",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	strebmClient := &gqltestutil.ComputeStrebmClient{Client: client}
	t.Run("strebm", func(t *testing.T) {
		testComputeClient(t, strebmClient)
	})
}
