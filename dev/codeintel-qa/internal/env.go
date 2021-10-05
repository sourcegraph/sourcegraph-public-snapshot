package internal

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	SourcegraphEndpoint    = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:3080", "Sourcegraph frontend endpoint")
	SourcegraphAccessToken = env.Get("SOURCEGRAPH_SUDO_TOKEN", "", "Sourcegraph access token with sudo privileges")
)
