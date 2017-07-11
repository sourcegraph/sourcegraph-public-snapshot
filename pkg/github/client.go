package github

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"context"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
)

var ghAppID, _ = strconv.Atoi(env.Get("SRC_GITHUB_APP_ID", "", "Integration ID for the Sourcegraph GitHub app."))
var ghAppKey = env.Get("SRC_GITHUB_APP_PRIVATE_KEY", "", "The private key for the Sourcegraph GitHub app.")

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

func githubConf(ctx context.Context) githubutil.Config {
	conf := *githubutil.Default
	conf.Context = ctx
	return conf
}

// Client returns the context's GitHub API client.
func Client(ctx context.Context) *github.Client {
	if MockRoundTripper != nil {
		return github.NewClient(&http.Client{
			Transport: MockRoundTripper,
		})
	}

	a := actor.FromContext(ctx)
	ghConf := githubConf(ctx)
	if a.GitHubToken != "" {
		return ghConf.AuthedClient(a.GitHubToken)
	}

	return ghConf.UnauthedClient()
}

func InstallationClient(ctx context.Context, installationID int) (*github.Client, error) {
	if MockRoundTripper != nil {
		return github.NewClient(&http.Client{
			Transport: MockRoundTripper,
		}), nil
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, ghAppID, installationID, []byte(ghAppKey))
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}

// UnauthedClient returns a github.Client that is unauthenticated
func UnauthedClient(ctx context.Context) *github.Client {
	if MockRoundTripper != nil {
		return github.NewClient(&http.Client{
			Transport: MockRoundTripper,
		})
	}

	conf := githubConf(ctx)
	return conf.UnauthedClient()
}

func checkResponse(ctx context.Context, resp *github.Response, err error, op string) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *github.RateLimitError:
		log15.Debug("exceeded GitHub rate limit", "error", err, "op", op)
		return err
	case *github.AbuseRateLimitError:
		log15.Debug("triggered GitHub abuse detection mechanism", "error", err, "op", op)
		abuseDetectionMechanismCounter.Inc()
		return err
	}

	if resp == nil {
		log15.Debug("no response from GitHub", "error", err)
		return err
	}

	switch resp.StatusCode {
	case 401, 404:
		// Pretty expected, not worth logging.
	case 451:
		log15.Debug("unavailable for legal reasons error received from GitHub", "error", err, "op", op)
		return err
	default:
		log15.Debug("unexpected error from GitHub", "error", err, "statusCode", resp.StatusCode, "op", op)
	}

	statusCode := errcode.HTTPToCode(resp.StatusCode)

	// Calling out to github could result in some HTTP status codes that don't directly map to
	// error status on sourcegraph. If github returns anything in the 400 range that isn't known to us,
	// we don't want to indicate a server-side error (which would happen if we don't convert
	// to 404 here).
	if statusCode == legacyerr.Unknown && resp.StatusCode >= 400 && resp.StatusCode < 500 {
		statusCode = legacyerr.NotFound
	}

	return legacyerr.Errorf(statusCode, "unexpected error from GitHub: %s: %v", op, err)
}

// HasAuthedUser reports whether the context has an authenticated
// GitHub user's credentials.
func HasAuthedUser(ctx context.Context) bool {
	return actor.FromContext(ctx).GitHubToken != ""
}

func IsRateLimitError(err error) bool {
	switch err.(type) {
	case *github.RateLimitError, *github.AbuseRateLimitError:
		return true
	default:
		return false
	}
}
