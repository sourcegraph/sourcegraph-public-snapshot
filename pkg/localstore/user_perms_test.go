package localstore

import (
	"context"
	"strings"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

// authTestContext with mock stubs for GitHubRepoGetter
func authTestContext() context.Context {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test", GitHubToken: "test"})
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")
	return ctx
}

func TestVerifyUserHasRepoURIAccess(t *testing.T) {
	ctx := authTestContext()

	tests := []struct {
		title                string
		repoURI              string
		authorizedForPrivate bool // True here simulates that the actor has access to private repos when asking GitHub API. False simulates that they don't.
		shouldCallGitHub     bool
		want                 bool
	}{
		{
			title:                `private repo URI begins with "github.com/", actor unauthorized for private repo access`,
			repoURI:              "github.com/user/privaterepo",
			authorizedForPrivate: false,
			shouldCallGitHub:     true,
			want:                 false,
		},
		{
			title:                `private repo URI begins with "GitHub.com/", actor unauthorized for private repo access`,
			repoURI:              "GitHub.com/User/PrivateRepo",
			authorizedForPrivate: false,
			shouldCallGitHub:     true,
			want:                 false,
		},
		{
			title:                `private repo URI begins with "github.com/", actor authorized for private repo access`,
			repoURI:              "github.com/user/privaterepo",
			authorizedForPrivate: true,
			shouldCallGitHub:     true,
			want:                 true,
		},
		{
			title:                `private repo URI begins with "GitHub.com/", actor authorized for private repo access`,
			repoURI:              "GitHub.com/User/PrivateRepo",
			authorizedForPrivate: true,
			shouldCallGitHub:     true,
			want:                 true,
		},
		{
			title:            `public repo URI begins with "github.com/"`,
			repoURI:          "github.com/user/publicrepo",
			shouldCallGitHub: true,
			want:             true,
		},
		{
			title:            `public repo URI begins with "GitHub.com/"`,
			repoURI:          "GitHub.com/User/PublicRepo",
			shouldCallGitHub: true,
			want:             true,
		},
		{
			title:            `repo URI begins with "bitbucket.org/"; not supported at this time, so permission denied`,
			repoURI:          "bitbucket.org/foo/bar",
			shouldCallGitHub: false,
			want:             false,
		},
		{
			title:            `repo URI that is local (neither GitHub nor a remote URI)`,
			repoURI:          "sourcegraph/sourcegraph",
			shouldCallGitHub: false,
			want:             false,
		},
	}
	for _, test := range tests {
		var calledGitHub = false

		// Mocked GitHub API responses differ depending on value of test.authorizedForPrivate.
		// If true, then "github.com/user/privaterepo" repo exists, otherwise it's not found.
		github.GetRepoMock = func(_ context.Context, uri string) (*sourcegraph.Repo, error) {
			calledGitHub = true
			switch uri := strings.ToLower(uri); {
			case uri == "github.com/user/privaterepo" && test.authorizedForPrivate:
				return &sourcegraph.Repo{URI: "github.com/User/PrivateRepo"}, nil
			case uri == "github.com/user/publicrepo":
				return &sourcegraph.Repo{URI: "github.com/User/PublicRepo"}, nil
			default:
				return nil, legacyerr.Errorf(legacyerr.NotFound, "repo not found")
			}
		}

		const repoID = 1
		got := verifyUserHasRepoURIAccess(ctx, test.repoURI)
		if calledGitHub != test.shouldCallGitHub {
			if test.shouldCallGitHub {
				t.Errorf("expected GitHub API to be called for permissions check, but it wasn't")
			} else {
				t.Errorf("did not expect GitHub API to be called for permissions check, but it was")
			}
		}
		if want := test.want; got != want {
			t.Errorf("%s: got %v, want %v", test.title, got, want)
		}
	}
}
