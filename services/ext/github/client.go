package github

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
)

var (
	abuseDetectionMechanismCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "github",
		Name:      "abuse_detection_mechanism",
		Help:      "Times that a response from GitHub indicated that abuse detection mechanism was triggered.",
	})
)

func init() {
	prometheus.MustRegister(abuseDetectionMechanismCounter)
}

var MockRoundTripper http.RoundTripper

// client returns the context's GitHub API client.
func client(ctx context.Context) *github.Client {
	if MockRoundTripper != nil {
		return github.NewClient(&http.Client{
			Transport: MockRoundTripper,
		})
	}

	ghConf := *githubutil.Default
	ghConf.Context = ctx

	a := auth.ActorFromContext(ctx)
	if a.GitHubToken != "" {
		return ghConf.AuthedClient(a.GitHubToken)
	}

	return ghConf.UnauthedClient()
}

func OrgsFromContext(ctx context.Context) *github.OrganizationsService {
	return client(ctx).Organizations
}

func checkResponse(ctx context.Context, resp *github.Response, err error, op string) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *github.RateLimitError:
		log15.Debug("exceeded github rate limit", "error", err, "op", op)
		return legacyerr.Errorf(legacyerr.ResourceExhausted, "exceeded GitHub API rate limit: %s: %v", op, err)
	case *github.AbuseRateLimitError:
		log15.Debug("triggered GitHub abuse detection mechanism", "error", err, "op", op)
		abuseDetectionMechanismCounter.Inc()
		return legacyerr.Errorf(legacyerr.ResourceExhausted, "triggered GitHub abuse detection mechanism: %s: %v", op, err)
	}

	if resp == nil {
		log15.Debug("no response from github", "error", err)
		return err
	}

	switch resp.StatusCode {
	case 401, 404:
		// Pretty expected, not worth logging.
	default:
		log15.Debug("unexpected error from github", "error", err, "statusCode", resp.StatusCode, "op", op)
	}

	statusCode := errcode.HTTPToCode(resp.StatusCode)

	// Calling out to github could result in some HTTP status codes that don't directly map to
	// error status on sourcegraph. If github returns anything in the 400 range that isn't known to us,
	// we don't want to indicate a server-side error (which would happen if we don't convert
	// to 404 here).
	if statusCode == legacyerr.Unknown && resp.StatusCode >= 400 && resp.StatusCode < 500 {
		statusCode = legacyerr.NotFound
	}

	return legacyerr.Errorf(statusCode, "%s", op)
}

// HasAuthedUser reports whether the context has an authenticated
// GitHub user's credentials.
func HasAuthedUser(ctx context.Context) bool {
	return auth.ActorFromContext(ctx).GitHubToken != ""
}

// HasClient returns true if the context currently contains a GitHub client. If no client exists in
// the context, `ReposFromContext` may still return a temporary one, but HasClient is useful for
// conditionals where we don't want to make GitHub requests unless a client has explicitly been set.
func HasClient(ctx context.Context) bool {
	client, _ := ctx.Value(minimalClientKey).(*minimalClient)
	repos, _ := ctx.Value(reposKey).(Repos)
	return client != nil || repos != nil
}
