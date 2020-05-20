package client

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var (
	preciseCodeIntelAPIServerURL = env.Get("PRECISE_CODE_INTEL_API_SERVER_URL", "k8s+http://precise-code-intel:3186", "precise-code-intel-api-server URL (or space separated list of precise-code-intel-api-server URLs)")

	preciseCodeIntelAPIServerURLsOnce sync.Once
	preciseCodeIntelAPIServerURLs     *endpoint.Map
)

type Client struct {
	endpoint   *endpoint.Map
	HTTPClient *http.Client
	server     *Server
}

func New(
	db db.DB,
	bundleManagerClient bundles.BundleManagerClient,
	codeIntelAPI api.CodeIntelAPI,
) *Client {
	return &Client{
		endpoint: LSIFURLs(),
		HTTPClient: &http.Client{
			// ot.Transport will propagate opentracing spans
			Transport: &ot.Transport{},
		},
		server: NewServer(db, bundleManagerClient, codeIntelAPI),
	}
}

func LSIFURLs() *endpoint.Map {
	preciseCodeIntelAPIServerURLsOnce.Do(func() {
		if len(strings.Fields(preciseCodeIntelAPIServerURL)) == 0 {
			preciseCodeIntelAPIServerURLs = endpoint.Empty(errors.New("an precise-code-intel-api-server has not been configured"))
		} else {
			preciseCodeIntelAPIServerURLs = endpoint.New(preciseCodeIntelAPIServerURL)
		}
	})
	return preciseCodeIntelAPIServerURLs
}
