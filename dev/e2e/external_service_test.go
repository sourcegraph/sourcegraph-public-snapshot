// +build e2e

package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/e2eutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestExternalService(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	t.Run("repositoryPathPattern", func(t *testing.T) {
		const repo = "sourcegraph/go-blame" // Tiny repo, fast to clone
		const slug = "github.com/" + repo
		// Set up external service
		esID, err := client.AddExternalService(e2eutil.AddExternalServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplayName: "e2e-test-github-repoPathPattern",
			Config: mustMarshalJSONString(struct {
				URL                   string   `json:"url"`
				Token                 string   `json:"token"`
				Repos                 []string `json:"repos"`
				RepositoryPathPattern string   `json:"repositoryPathPattern"`
			}{
				URL:                   "http://github.com",
				Token:                 *githubToken,
				Repos:                 []string{repo},
				RepositoryPathPattern: "foobar/{host}/{nameWithOwner}",
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

		err = client.WaitForReposToBeCloned("foobar/" + slug)
		if err != nil {
			t.Fatal(err)
		}

		// The request URL should be redirected to the new path
		origURL := *baseURL + "/" + slug
		resp, err := client.Get(origURL)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		wantURL := *baseURL + "/foobar/" + slug // <baseURL>/foobar/github.com/sourcegraph/go-blame
		if diff := cmp.Diff(wantURL, resp.Request.URL.String()); diff != "" {
			t.Fatalf("URL mismatch (-want +got):\n%s", diff)
		}
	})
}
