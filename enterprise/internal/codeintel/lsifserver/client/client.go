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
	lsifURL = env.Get("LSIF_API_URL", "k8s+http://lsif-api:3186", "lsif-api URL (or space separated list of lsif-api URLs)")

	lsifURLsOnce sync.Once
	lsifURLs     *endpoint.Map

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
	lsifURLsOnce.Do(func() {
		if len(strings.Fields(lsifURL)) == 0 {
			lsifURLs = endpoint.Empty(errors.New("an lsif-api has not been configured"))
		} else {
			lsifURLs = endpoint.New(lsifURL)
		}
	})
	return lsifURLs
}
