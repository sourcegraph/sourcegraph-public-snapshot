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
	// TODO - rename me
	lsifServerURL = env.Get("LSIF_API_URL", "k8s+http://lsif-api:3186", "lsif-api URL (or space separated list of lsif-server URLs)")

	lsifServerURLsOnce sync.Once
	lsifServerURLs     *endpoint.Map

	DefaultClient = &Client{
		endpoint: LSIFServerURLs(),
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

func LSIFServerURLs() *endpoint.Map {
	lsifServerURLsOnce.Do(func() {
		if len(strings.Fields(lsifServerURL)) == 0 {
			lsifServerURLs = endpoint.Empty(errors.New("an lsif-server has not been configured"))
		} else {
			lsifServerURLs = endpoint.New(lsifServerURL)
		}
	})
	return lsifServerURLs
}
