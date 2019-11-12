package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintelligence/lsifserver"
)

type LSIFServerProxy struct {
	UploadHandler    http.Handler
	AllRoutesHandler http.Handler
}

func NewProxy() (*LSIFServerProxy, error) {
	url, err := url.Parse(lsifserver.ServerURLFromEnv)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return &LSIFServerProxy{
		UploadHandler:    http.HandlerFunc(uploadProxyHandler(proxy)),
		AllRoutesHandler: http.HandlerFunc(allRoutesProxyHandler(proxy)),
	}, nil
}
