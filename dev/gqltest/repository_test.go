package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestRepository(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-repository",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("external code host links", func(t *testing.T) {
		got, err := client.FileExternalLinks(
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
	if *azureDevOpsUsername == "" || *azureDevOpsToken == "" {
		t.Skip("Environment variable AZURE_DEVOPS_USERNAME or AZURE_DEVOPS_TOKEN is not set")
	}

	t.Skip("Test Repo is gone from Azure Devops and only admins can create repos. SQS is on vacation and he's the only admin. We don't know how this repo got deleted.")

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindOther,
		DisplayName: "gqltest-azure-devops-repository",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL: fmt.Sprintf("https://%s:%s@sourcegraph.visualstudio.com/sourcegraph/_git/", *azureDevOpsUsername, *azureDevOpsToken),
			Repos: []string{
				"Test Repo",
			},
			RepositoryPathPattern: "sourcegraph.visualstudio.com/{repo}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	err = client.WaitForReposToBeCloned(
		"sourcegraph.visualstudio.com/Test Repo",
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.Repository("sourcegraph.visualstudio.com/Test Repo")
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
