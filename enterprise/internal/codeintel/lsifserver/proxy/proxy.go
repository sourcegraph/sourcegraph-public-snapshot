package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver"
)

func NewProxy() (*httpapi.LSIFServerProxy, error) {
	url, err := url.Parse(lsifserver.ServerURLFromEnv)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return &httpapi.LSIFServerProxy{
		UploadHandler:    http.HandlerFunc(uploadProxyHandler(proxy)),
		AllRoutesHandler: http.HandlerFunc(allRoutesProxyHandler(proxy)),
	}, nil
}
