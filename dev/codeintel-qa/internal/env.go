package internal

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	// Ensure required environment variables are set
	SourcegraphEndpoint    = env.Get("SRC_ENDPOINT", "http://127.0.0.1:3080", "Sourcegraph frontend endpoint")
	SourcegraphAccessToken = env.Get("SRC_ACCESS_TOKEN", "", "Sourcegraph access token with sudo privileges")
)
