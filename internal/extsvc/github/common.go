package github

import (
	"log"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

var (
	gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degradation"))
	githubProxyURL   = func() *url.URL {
		url, err := url.Parse(env.Get("GITHUB_BASE_URL", "http://github-proxy", "base URL for GitHub.com API (used for github-proxy)"))
		if err != nil {
			log.Fatal("Error parsing GITHUB_BASE_URL:", err)
		}
		return url
	}()

	requestCounter = metrics.NewRequestMeter("github", "Total number of requests sent to the GitHub API.")
)
