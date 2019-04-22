package repos

import (
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetGitHubConnection(t *testing.T) {
	orig := githubConnections.Get()
	githubConnections.Set(func() interface{} {
		return []*githubConnection{
			{originalHostname: "github.com", baseURL: &url.URL{Scheme: "https", Host: "github.com", Path: "/"}, config: &schema.GitHubConnection{Token: "t"}},
			{originalHostname: "github.example.com", baseURL: &url.URL{Scheme: "https", Host: "github.example.com", Path: "/"}, config: &schema.GitHubConnection{Token: "t"}},
		}
	})
	defer func() { githubConnections.Set(func() interface{} { return orig }) }()

	githubConnections := githubConnections.Get().([]*githubConnection)
	t.Run("not github", func(t *testing.T) {
		c, err := getGitHubConnection(protocol.RepoLookupArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if c != nil {
			t.Errorf("got conn %+v, want nil", c)
		}
	})

	t.Run("by repo name", func(t *testing.T) {
		t.Run("github.com", func(t *testing.T) {
			c, err := getGitHubConnection(protocol.RepoLookupArgs{Repo: "github.com/foo/bar"})
			if err != nil {
				t.Fatal(err)
			}
			if want := githubConnections[0]; c != want {
				t.Errorf("got conn %+v, want %+v", c, want)
			}
		})

		t.Run("github.example.com", func(t *testing.T) {
			c, err := getGitHubConnection(protocol.RepoLookupArgs{Repo: "github.example.com/foo/bar"})
			if err != nil {
				t.Fatal(err)
			}
			if want := githubConnections[1]; c != want {
				t.Errorf("got conn %+v, want %+v", c, want)
			}
		})
	})

	t.Run("by external repository spec", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := getGitHubConnection(protocol.RepoLookupArgs{ExternalRepo: &api.ExternalRepoSpec{ServiceType: github.ServiceType, ServiceID: "https://github.is-not-configured.com/"}})
			if err == nil {
				t.Fatal("err == nil")
			}
		})

		t.Run("github.com", func(t *testing.T) {
			c, err := getGitHubConnection(protocol.RepoLookupArgs{ExternalRepo: &api.ExternalRepoSpec{ServiceType: github.ServiceType, ServiceID: "https://github.com/"}})
			if err != nil {
				t.Fatal(err)
			}
			if want := githubConnections[0]; c != want {
				t.Errorf("got conn %+v, want %+v", c, want)
			}
		})

		t.Run("github.example.com", func(t *testing.T) {
			c, err := getGitHubConnection(protocol.RepoLookupArgs{ExternalRepo: &api.ExternalRepoSpec{ServiceType: github.ServiceType, ServiceID: "https://github.example.com/"}})
			if err != nil {
				t.Fatal(err)
			}
			if want := githubConnections[1]; c != want {
				t.Errorf("got conn %+v, want %+v", c, want)
			}
		})
	})
}

func TestExampleRepositoryQuerySplit(t *testing.T) {
	q := "org:sourcegraph"
	want := `["org:sourcegraph created:>=2019","org:sourcegraph created:2018","org:sourcegraph created:2016..2017","org:sourcegraph created:<2016"]`
	have := exampleRepositoryQuerySplit(q)
	if want != have {
		t.Errorf("unexpected example query for %s:\nwant: %s\nhave: %s", q, want, have)
	}
}
