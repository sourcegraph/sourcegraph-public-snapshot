// +build gqltest

package main

import (
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
	defer func() {
		err := client.DeleteExternalService(esID)
		if err != nil {
			t.Fatal(err)
		}
	}()

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
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}
