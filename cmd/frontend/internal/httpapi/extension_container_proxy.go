package httpapi

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func extensionContainerProxy(w http.ResponseWriter, r *http.Request) {
	extensionContainerProxyHandlerMu.Lock()
	h := extensionContainerProxyHandler
	extensionContainerProxyHandlerMu.Unlock()
	if h != nil {
		h.ServeHTTP(w, r)
	} else {
		http.Error(w, "extension container proxy is not yet ready", http.StatusTooEarly)
	}
}

var (
	extensionContainerProxyHandlerMu sync.Mutex
	extensionContainerProxyHandler   http.Handler
)

func init() {
	var lastValue []*schema.ExtensionsContainers
	go func() {
		conf.Watch(func() {
			nextValue := conf.Get().ExtensionsContainers
			if reflect.DeepEqual(lastValue, nextValue) {
				return // no change
			}

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "no extension container proxy for this path", http.StatusNotFound)
			})
			mappings := parseExtensionContainerMappings(nextValue)
			for path, toURL := range mappings {
				mux.Handle(path, http.StripPrefix(path, newExtensionContainerReverseProxy(toURL)))
			}

			extensionContainerProxyHandlerMu.Lock()
			defer extensionContainerProxyHandlerMu.Unlock()
			extensionContainerProxyHandler = mux
		})
	}()
}

func newExtensionContainerReverseProxy(toURL url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = toURL.ResolveReference(&url.URL{
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
			})
			r.Host = toURL.Host
			r.Header.Del("Authorization")
			r.Header.Del("Cookie")
			// TODO!(sqs): whitelist cookies and headers
		},
		ModifyResponse: func(resp *http.Response) error {
			// TODO!(sqs): whitelist cookies and headers
			resp.Header.Del("Set-Cookie")
			return nil
		},
	}
}

func parseExtensionContainerMappings(rawMappings []*schema.ExtensionsContainers) map[string]url.URL {
	mappings := make(map[string]url.URL, len(rawMappings))
	for _, m := range rawMappings {
		fromPath := m.From
		const wantPrefix = "/.api/extension-containers/"
		if !strings.HasPrefix(fromPath, wantPrefix) {
			log15.Warn("Invalid extensions.containers configuration value (from path must begin with prefix).", "from", m.From, "requiredPrefix", wantPrefix)
			continue
		}
		fromPath = strings.TrimPrefix(fromPath, wantPrefix[:len(wantPrefix)-1]) // preserve leading slash
		fromPath = strings.TrimSuffix(fromPath, "/")

		toURL, err := url.Parse(m.To)
		if err != nil {
			log15.Warn("Invalid extensions.containers configuration value (to URL is not a valid URL).", "to", m.To)
			continue
		}
		if otherURL, exists := mappings[fromPath]; exists {
			log15.Warn("Duplicate mappings in extensions.containers configuration value. Behavior is undefined.", "from", fromPath, "to1", toURL.String(), "to2", otherURL.String())
			continue
		}

		mappings[fromPath] = *toURL
	}
	return mappings
}
