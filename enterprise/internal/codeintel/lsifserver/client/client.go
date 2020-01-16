package client

import (
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver"
)

var DefaultClient = &Client{
	URL: lsifserver.ServerURLFromEnv,
	HTTPClient: &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{},
	},
}

type Client struct {
	URL        string
	HTTPClient *http.Client
}
