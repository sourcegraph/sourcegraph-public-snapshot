package github

import (
	"net/http"

	"github.com/sourcegraph/go-github/github"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
)

// minimalClient contains the minimal set of GitHub API methods needed
// by this package.
type minimalClient struct {
	repos githubRepos
	orgs  githubOrgs

	isAuthedUser bool // whether the client is using a GitHub user's auth token
}

func newMinimalClient(client *github.Client, isAuthedUser bool) *minimalClient {
	return &minimalClient{
		repos: client.Repositories,
		orgs:  client.Organizations,

		isAuthedUser: isAuthedUser,
	}
}

type githubRepos interface {
	Get(owner, repo string) (*github.Repository, *github.Response, error)
	GetByID(id int) (*github.Repository, *github.Response, error)
	List(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error)
}

type githubOrgs interface {
	ListMembers(org string, opt *github.ListMembersOptions) ([]github.User, *github.Response, error)
	List(member string, opt *github.ListOptions) ([]github.Organization, *github.Response, error)
}

func checkResponse(resp *github.Response, err error, op string) error {
	if err == nil {
		return nil
	}
	if resp == nil {
		log15.Debug("no response from github", "error", err)
		return err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		log15.Debug("unexpected error from github", "error", err, "statusCode", resp.StatusCode, "op", op)
	}

	statusCode := errcode.HTTPToGRPC(resp.StatusCode)

	// Calling out to github could result in some HTTP status codes that don't directly map to
	// gRPC status codes. If github returns anything in the 400 range that isn't known to us,
	// we don't want to indicate a server-side error (which would happen if we don't convert
	// to 404 here).
	if statusCode == codes.Unknown && resp.StatusCode >= 400 && resp.StatusCode < 500 {
		statusCode = codes.NotFound
	}

	return grpc.Errorf(statusCode, "%s", op)
}
