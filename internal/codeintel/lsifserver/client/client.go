package client

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var (
	preciseCodeIntelServerURL = env.Get("PRECISE_CODE_INTEL_SERVER_URL", "k8s+http://precise-code-intel:3186", "precise-code-intel-server URL (or space separated list of precise-code-intel-server URLs)")

	preciseCodeIntelServerURLsOnce sync.Once
	preciseCodeIntelServerURLs     *endpoint.Map

	DefaultClient = &Client{
		endpoint: LSIFURLs(),
		HTTPClient: &http.Client{
			// ot.Transport will propagate opentracing spans
			Transport: &ot.Transport{},
		},
	}
)

type Client struct {
	endpoint   *endpoint.Map
	HTTPClient *http.Client
}

func LSIFURLs() *endpoint.Map {
	preciseCodeIntelServerURLsOnce.Do(func() {
		if len(strings.Fields(preciseCodeIntelServerURL)) == 0 {
			preciseCodeIntelServerURLs = endpoint.Empty(errors.New("an precise-code-intel-server has not been configured"))
		} else {
			preciseCodeIntelServerURLs = endpoint.New(preciseCodeIntelServerURL)
		}
	})
	return preciseCodeIntelServerURLs
}
