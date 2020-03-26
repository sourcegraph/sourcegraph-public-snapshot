package client

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	// TODO - update this to lsif-api-server once we split them in k8s
	lsifAPIServerURL = env.Get("LSIF_API_SERVER_URL", "k8s+http://lsif-server:3186", "lsif-api-server URL (or space separated list of lsif-api-server URLs)")

	lsifAPIServerURLsOnce sync.Once
	lsifAPIServerURLs     *endpoint.Map

	DefaultClient = &Client{
		endpoint: LSIFURLs(),
		HTTPClient: &http.Client{
			// nethttp.Transport will propagate opentracing spans
			Transport: &nethttp.Transport{},
		},
	}
)

type Client struct {
	endpoint   *endpoint.Map
	HTTPClient *http.Client
}

func LSIFURLs() *endpoint.Map {
	lsifAPIServerURLsOnce.Do(func() {
		if len(strings.Fields(lsifAPIServerURL)) == 0 {
			lsifAPIServerURLs = endpoint.Empty(errors.New("an lsif-api-server has not been configured"))
		} else {
			lsifAPIServerURLs = endpoint.New(lsifAPIServerURL)
		}
	})
	return lsifAPIServerURLs
}
