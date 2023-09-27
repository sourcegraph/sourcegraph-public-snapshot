pbckbge mbin

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

func TestRepository(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment vbribble GITHUB_TOKEN is not set")
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "gqltest-github-repository",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
		}),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	err = client.WbitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("externbl code host links", func(t *testing.T) {
		got, err := client.FileExternblLinks(
			"github.com/sgtest/go-diff",
			"3f415b150bec0685cb81b73cc201e762e075006d",
			"diff/pbrse.go",
		)
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := []*gqltestutil.ExternblLink{
			{
				URL:         "https://ghe.sgdev.org/sgtest/go-diff/blob/3f415b150bec0685cb81b73cc201e762e075006d/diff/pbrse.go",
				ServiceType: extsvc.TypeGitHub,
				ServiceKind: extsvc.KindGitHub,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestRepository_NbmeWithSpbce(t *testing.T) {
	if *bzureDevOpsUsernbme == "" || *bzureDevOpsToken == "" {
		t.Skip("Environment vbribble AZURE_DEVOPS_USERNAME or AZURE_DEVOPS_TOKEN is not set")
	}

	t.Skip("Test Repo is gone from Azure Devops bnd only bdmins cbn crebte repos. SQS is on vbcbtion bnd he's the only bdmin. We don't know how this repo got deleted.")

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindOther,
		DisplbyNbme: "gqltest-bzure-devops-repository",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL: fmt.Sprintf("https://%s:%s@sourcegrbph.visublstudio.com/sourcegrbph/_git/", *bzureDevOpsUsernbme, *bzureDevOpsToken),
			Repos: []string{
				"Test Repo",
			},
			RepositoryPbthPbttern: "sourcegrbph.visublstudio.com/{repo}",
		}),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	err = client.WbitForReposToBeCloned(
		"sourcegrbph.visublstudio.com/Test Repo",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	got, err := client.Repository("sourcegrbph.visublstudio.com/Test Repo")
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &gqltestutil.Repository{
		ID:  got.ID,
		URL: "/sourcegrbph.visublstudio.com/Test%20Repo",
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}
