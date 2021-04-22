package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAWSCodeCloneURLs(t *testing.T) {
	clock := timeutil.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	repo := &awscodecommit.Repository{
		ARN:          "arn:aws:codecommit:us-west-1:999999999999:stripe-go",
		AccountID:    "999999999999",
		ID:           "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
		Name:         "stripe-go",
		Description:  "The stripe-go lib",
		HTTPCloneURL: "https://git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go",
		LastModified: &now,
	}

	cfg := schema.AWSCodeCommitConnection{
		GitCredentials: schema.AWSCodeCommitGitCredentials{
			Username: "username",
			Password: "password",
		},
	}

	got := awsCodeCloneURL(repo, &cfg)
	want := "https://username:password@git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go"
	if got != want {
		t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
	}
}

func TestBitbucketServerCloneURLs(t *testing.T) {
	repo := &bitbucketserver.Repo{
		ID:   1,
		Slug: "bar",
		Project: &bitbucketserver.Project{
			Key: "foo",
		},
	}

	cfg := schema.BitbucketServerConnection{
		Token:    "abc",
		Username: "username",
		Password: "password",
	}

	t.Run("ssh", func(t *testing.T) {
		repo.Links.Clone = []struct {
			Href string "json:\"href\""
			Name string "json:\"name\""
		}{
			// even if the first link is http, ssh should prevail
			{Name: "http", Href: "https://asdine@bitbucket.example.com/scm/sg/sourcegraph.git"},
			{Name: "ssh", Href: "ssh://git@bitbucket.example.com:7999/sg/sourcegraph.git"},
		}

		cfg.GitURLType = "ssh" // use ssh in the config as well

		got := bitbucketServerCloneURL(repo, &cfg)
		want := "ssh://git@bitbucket.example.com:7999/sg/sourcegraph.git"
		if got != want {
			t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
		}
	})

	t.Run("http", func(t *testing.T) {
		// Second test: http
		repo.Links.Clone = []struct {
			Href string "json:\"href\""
			Name string "json:\"name\""
		}{
			{Name: "http", Href: "https://asdine@bitbucket.example.com/scm/sg/sourcegraph.git"},
		}

		got := bitbucketServerCloneURL(repo, &cfg)
		want := "https://asdine:abc@bitbucket.example.com/scm/sg/sourcegraph.git"
		if got != want {
			t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
		}
	})

	t.Run("no token", func(t *testing.T) {
		// Third test: no token
		cfg.Token = ""

		got := bitbucketServerCloneURL(repo, &cfg)
		want := "https://asdine:password@bitbucket.example.com/scm/sg/sourcegraph.git"
		if got != want {
			t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
		}
	})
}

func TestBitbucketCloudCloneURLs(t *testing.T) {
	repo := &bitbucketcloud.Repo{
		FullName: "sg/sourcegraph",
	}

	repo.Links.Clone = []struct {
		Href string "json:\"href\""
		Name string "json:\"name\""
	}{
		{Name: "https", Href: "https://asdine@bitbucket.org/sg/sourcegraph.git"},
		{Name: "ssh", Href: "git@bitbucket.org/sg/sourcegraph.git"},
	}

	cfg := schema.BitbucketCloudConnection{
		Url:         "bitbucket.org",
		Username:    "username",
		AppPassword: "password",
	}

	t.Run("ssh", func(t *testing.T) {
		cfg.GitURLType = "ssh"

		got := bitbucketCloudCloneURL(repo, &cfg)
		want := "git@bitbucket.org:sg/sourcegraph.git"
		if got != want {
			t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
		}
	})

	t.Run("http", func(t *testing.T) {
		cfg.GitURLType = "http"

		got := bitbucketCloudCloneURL(repo, &cfg)
		want := "https://username:password@bitbucket.org/sg/sourcegraph.git"
		if got != want {
			t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
		}
	})
}

func TestGithubCloneURLs(t *testing.T) {
	var repo github.Repository
	repo.NameWithOwner = "foo/bar"

	tests := []struct {
		InstanceUrl string
		RepoURL     string
		Token       string
		GitURLType  string
		Want        string
	}{
		{"https://github.com", "https://github.com/foo/bar", "", "", "https://github.com/foo/bar"},
		{"https://github.com", "https://github.com/foo/bar", "abcd", "", "https://abcd@github.com/foo/bar"},
		{"https://github.com", "https://github.com/foo/bar", "abcd", "ssh", "git@github.com:foo/bar.git"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("URL(%q) / Token(%q) / URLType(%q)", test.InstanceUrl, test.Token, test.GitURLType), func(t *testing.T) {
			cfg := schema.GitHubConnection{
				Url:        test.InstanceUrl,
				Token:      test.Token,
				GitURLType: test.GitURLType,
			}

			repo.URL = test.RepoURL

			got, err := githubCloneURL(&repo, &cfg)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.Want {
				t.Fatalf("wrong cloneURL, got: %q, want: %q", got, test.Want)
			}
		})
	}
}

func TestGitlabCloneURLs(t *testing.T) {
	repo := &gitlab.Project{
		ProjectCommon: gitlab.ProjectCommon{
			ID:                1,
			PathWithNamespace: "foo/bar",
			SSHURLToRepo:      "git@gitlab.com:gitlab-org/gitaly.git",
			HTTPURLToRepo:     "https://gitlab.com/gitlab-org/gitaly.git",
		},
	}

	tests := []struct {
		Token      string
		GitURLType string
		Want       string
	}{
		{"", "", "https://gitlab.com/gitlab-org/gitaly.git"},
		{"abcd", "", "https://git:abcd@gitlab.com/gitlab-org/gitaly.git"},
		{"abcd", "ssh", "git@gitlab.com:gitlab-org/gitaly.git"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Token(%q) / URLType(%q)", test.Token, test.GitURLType), func(t *testing.T) {
			cfg := schema.GitLabConnection{
				Token:      test.Token,
				GitURLType: test.GitURLType,
			}

			got := gitlabCloneURL(repo, &cfg)
			if got != test.Want {
				t.Fatalf("wrong cloneURL, got: %q, want: %q", got, test.Want)
			}
		})
	}
}

func TestPerforceCloneURL(t *testing.T) {
	cfg := schema.PerforceConnection{
		P4Port:   "ssl:111.222.333.444:1666",
		P4User:   "admin",
		P4Passwd: "pa$$word",
	}

	repo := &perforce.Depot{
		Depot: "//Sourcegraph/",
	}

	got := perforceCloneURL(repo, &cfg)
	want := "perforce://admin:pa$$word@ssl:111.222.333.444:1666//Sourcegraph/"
	if got != want {
		t.Fatalf("wrong cloneURL, got: %q, want: %q", got, want)
	}
}

func TestSetUserinfoBestEffort(t *testing.T) {
	cases := []struct {
		rawurl   string
		username string
		password string
		want     string
	}{
		// no-op
		{"https://foo.com/foo/bar", "", "", "https://foo.com/foo/bar"},
		// invalid name is returned as is
		{":/foo.com/foo/bar", "u", "p", ":/foo.com/foo/bar"},

		// no user details in rawurl
		{"https://foo.com/foo/bar", "u", "p", "https://u:p@foo.com/foo/bar"},
		{"https://foo.com/foo/bar", "u", "", "https://u@foo.com/foo/bar"},
		{"https://foo.com/foo/bar", "", "p", "https://foo.com/foo/bar"},

		// user set already
		{"https://x@foo.com/foo/bar", "u", "p", "https://x:p@foo.com/foo/bar"},
		{"https://x@foo.com/foo/bar", "u", "", "https://x@foo.com/foo/bar"},
		{"https://x@foo.com/foo/bar", "", "p", "https://x:p@foo.com/foo/bar"},

		// user and password set already
		{"https://x:y@foo.com/foo/bar", "u", "p", "https://x:y@foo.com/foo/bar"},
		{"https://x:y@foo.com/foo/bar", "u", "", "https://x:y@foo.com/foo/bar"},
		{"https://x:y@foo.com/foo/bar", "", "p", "https://x:y@foo.com/foo/bar"},

		// empty password
		{"https://x:@foo.com/foo/bar", "u", "p", "https://x:@foo.com/foo/bar"},
		{"https://x:@foo.com/foo/bar", "u", "", "https://x:@foo.com/foo/bar"},
		{"https://x:@foo.com/foo/bar", "", "p", "https://x:@foo.com/foo/bar"},
	}
	for _, c := range cases {
		got := setUserinfoBestEffort(c.rawurl, c.username, c.password)
		if got != c.want {
			t.Errorf("setUserinfoBestEffort(%q, %q, %q): got %q want %q", c.rawurl, c.username, c.password, got, c.want)
		}
	}
}
