package github

import (
	"net/http"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
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

	ghConf := githubConf(ctx)
	return ghConf.UnauthedClient()
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
