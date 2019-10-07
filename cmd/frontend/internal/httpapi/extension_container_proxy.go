package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var extensionContainerProxyHandler http.Handler

var extensionContainers = env.Get("EXTENSION_CONTAINERS", "", "a space-separated list of proxy mappings from URL path to external URL, such as: /.api/extension-containers/FOO->http://FOO:1234")

func init() {
	mappings, err := parseExtensionContainerMappings(extensionContainers)
	if err != nil {
		log.Fatalf("EXTENSION_CONTAINERS: %s", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("PROXY %q", r.URL)
	})
	for path, toURL := range mappings {
		log.Printf("MAP %q -> %q", path, toURL)
		mux.Handle(path, http.StripPrefix(path, newExtensionContainerReverseProxy(toURL)))
	}
	extensionContainerProxyHandler = mux
}

func newExtensionContainerReverseProxy(toURL url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = toURL.ResolveReference(&url.URL{
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
			})
			r.Host = toURL.Host
		},
	}
}

func parseExtensionContainerMappings(mappingsStr string) (map[string]url.URL, error) {
	mappingsStrs := strings.Fields(mappingsStr)
	mappings := make(map[string]url.URL, len(mappingsStrs))
	for _, s := range mappingsStrs {
		parts := strings.Split(s, "->")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid extension containers mapping entry %q (does not contain \"->\")", s)
		}

		fromPath := parts[0]
		const wantPrefix = "/.api/extension-containers/"
		if !strings.HasPrefix(fromPath, wantPrefix) {
			return nil, fmt.Errorf("invalid extension containers mapping entry %q (path must begin with %q)", s, wantPrefix)
		}
		fromPath = strings.TrimPrefix(fromPath, wantPrefix[:len(wantPrefix)-1]) // preserve leading slash
		fromPath = strings.TrimSuffix(fromPath, "/")

		toURL, err := url.Parse(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid extension containers mapping entry %q (%s)", s, err)
		}
		if otherURL, exists := mappings[fromPath]; exists {
			return nil, fmt.Errorf("duplicate extension containers mapping entries for path %q (URLs: %q %q)", fromPath, toURL, otherURL)
		}

		mappings[fromPath] = *toURL
	}
	return mappings, nil
}
