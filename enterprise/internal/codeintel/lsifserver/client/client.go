package client

import (
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
)

var DefaultClient = &Client{
	endpoint: lsifserver.LSIFServerURLs(),
	HTTPClient: &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{},
	},
}

type Client struct {
	endpoint   *endpoint.Map
	HTTPClient *http.Client
}
