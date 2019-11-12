package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintelligence"
)

type LSIFProxy struct {
	ProxyHandler  http.Handler
	UploadHandler http.Handler
}

func NewProxy() (*LSIFProxy, error) {
	url, err := url.Parse(codeintelligence.ServerURLFromEnv)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return &LSIFProxy{
		ProxyHandler:  http.HandlerFunc(ProxyHandler(proxy)),
		UploadHandler: http.HandlerFunc(UploadProxyHandler(proxy)),
	}, nil
}
