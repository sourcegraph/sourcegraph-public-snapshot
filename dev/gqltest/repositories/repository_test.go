package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/gqltest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepository(t *testing.T) {
	if len(*gqltest.GithubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := gqltest.Client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-repository",
		Config: gqltest.MustMarshalJSONString(&schema.GitHubConnection{
			Url:   "https://ghe.sgdev.org/",
			Token: *gqltest.GithubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	gqltest.RemoveExternalServiceAfterTest(t, esID)

	err = gqltest.Client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("external code host links", func(t *testing.T) {
		got, err := gqltest.Client.FileExternalLinks(
			"github.com/sgtest/go-diff",
			"3f415a150aec0685cb81b73cc201e762e075006d",
			"diff/parse.go",
		)
		if err != nil {
			t.Fatal(err)
		}

		want := []*gqltestutil.ExternalLink{
			{
				URL:         "https://ghe.sgdev.org/sgtest/go-diff/blob/3f415a150aec0685cb81b73cc201e762e075006d/diff/parse.go",
				ServiceType: extsvc.TypeGitHub,
				ServiceKind: extsvc.KindGitHub,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestRepository_NameWithSpace(t *testing.T) {
	if *gqltest.AzureDevOpsUsername == "" || *gqltest.AzureDevOpsToken == "" {
		t.Skip("Environment variable AZURE_DEVOPS_USERNAME or AZURE_DEVOPS_TOKEN is not set")
	}

	t.Skip("Test Repo is gone from Azure Devops and only admins can create repos. SQS is on vacation and he's the only admin. We don't know how this repo got deleted.")

	// Set up external service
	esID, err := gqltest.Client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindOther,
		DisplayName: "gqltest-azure-devops-repository",
		Config: gqltest.MustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL: fmt.Sprintf("https://%s:%s@sourcegraph.visualstudio.com/sourcegraph/_git/", *gqltest.AzureDevOpsUsername, *gqltest.AzureDevOpsToken),
			Repos: []string{
				"Test Repo",
			},
			RepositoryPathPattern: "sourcegraph.visualstudio.com/{repo}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	gqltest.RemoveExternalServiceAfterTest(t, esID)

	err = gqltest.Client.WaitForReposToBeCloned(
		"sourcegraph.visualstudio.com/Test Repo",
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := gqltest.Client.Repository("sourcegraph.visualstudio.com/Test Repo")
	if err != nil {
		t.Fatal(err)
	}

	want := &gqltestutil.Repository{
		ID:  got.ID,
		URL: "/sourcegraph.visualstudio.com/Test%20Repo",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}
