package github

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
)

var rateLimitRemainingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
	Help:      "Number of calls to GitHub's API remaining before hitting the rate limit.",
})

func init() {
	rateLimitRemainingGauge.Set(5000)
	prometheus.MustRegister(rateLimitRemainingGauge)
}

// minimalClient contains the minimal set of GitHub API methods needed
// by this package.
type minimalClient struct {
	repos githubRepos
	orgs  githubOrgs

	// These are authenticated as the OAuth2 client application using
	// HTTP Basic auth, not as the user. (Some GitHub API endpoints
	// require that.)
	appAuthorizations githubAuthorizations

	isAuthedUser bool // whether the client is using a GitHub user's auth token
}

func newMinimalClient(isAuthedUser bool, userClient *github.Client, appClient *github.Client) *minimalClient {
	return &minimalClient{
		repos: userClient.Repositories,
		orgs:  userClient.Organizations,

		appAuthorizations: appClient.Authorizations,

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

type githubAuthorizations interface {
	Revoke(clientID, token string) (*github.Response, error)
}

func checkResponse(ctx context.Context, resp *github.Response, err error, op string) error {
	if err == nil {
		return nil
	}
	if resp == nil {
		log15.Debug("no response from github", "error", err)
		return err
	}

	rateLimitRemainingGauge.Set(float64(resp.Remaining))

	if _, ok := err.(*github.RateLimitError); ok {
		log15.Debug("exceeded github rate limit", "error", err, "op", op)
		return grpc.Errorf(codes.ResourceExhausted, "exceeded GitHub API rate limit: %s: %v", op, err)
	}
	if uid := auth.ActorFromContext(ctx).UID; resp.StatusCode == http.StatusUnauthorized && uid != 0 {
		// Token possibly revoked, so delete it from DB.
		_, err = svc.Auth(ctx).DeleteAndRevokeExternalToken(ctx, &sourcegraph.ExternalTokenSpec{
			UID: int32(uid),
		})
		if err != nil {
			log15.Error("could not delete external token", "error", err, "uid", uid)
		}
	} else {
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
