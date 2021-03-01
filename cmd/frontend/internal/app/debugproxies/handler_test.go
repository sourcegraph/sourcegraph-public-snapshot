package debugproxies

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
)

func TestReverseProxyRequestPaths(t *testing.T) {
	var rph ReverseProxyHandler

	proxiedServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Path))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Parse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displayName := displayNameFromEndpoint(ep)
	rph.Populate([]Endpoint{ep})

	ctx := backend.WithAuthzBypass(context.Background())

	link := fmt.Sprintf("%s/-/debug/proxies/%s/metrics", proxiedServer.URL, displayName)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	rtr := mux.NewRouter()
	rtr.PathPrefix("/-/debug").Name(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter())

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if string(body) != "/metrics" {
		t.Errorf("expected /metrics to be passed to reverse proxy, got %s", body)
	}
}

func TestIndexLinks(t *testing.T) {
	var rph ReverseProxyHandler

	proxiedServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Path))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Parse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displayName := displayNameFromEndpoint(ep)
	rph.Populate([]Endpoint{ep})

	ctx := backend.WithAuthzBypass(context.Background())

	link := fmt.Sprintf("%s/-/debug/", proxiedServer.URL)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	rtr := mux.NewRouter()
	rtr.PathPrefix("/-/debug").Name(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter())

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	expectedContent := fmt.Sprintf("<a href=\"proxies/%s/\">%s</a><br>", displayName, displayName)

	if !strings.Contains(string(body), expectedContent) {
		t.Errorf("expected %s, got %s", expectedContent, body)
	}
}

func TestDisplayNameFromEndpoint(t *testing.T) {
	cases := []struct {
		Service, Addr, Hostname string
		Want                    string
	}{{
		Service:  "gitserver",
		Addr:     "192.168.10.0:2323",
		Hostname: "gitserver-0",
		Want:     "gitserver-0",
	}, {
		Service: "searcher",
		Addr:    "192.168.10.3:2323",
		Want:    "searcher-192.168.10.3",
	}, {
		Service: "no-port",
		Addr:    "192.168.10.1",
		Want:    "no-port-192.168.10.1",
	}}

	for _, c := range cases {
		got := displayNameFromEndpoint(Endpoint{
			Service:  c.Service,
			Addr:     c.Addr,
			Hostname: c.Hostname,
		})
		if got != c.Want {
			t.Errorf("displayNameFromEndpoint(%q, %q) mismatch (-want +got):\n%s", c.Service, c.Addr, cmp.Diff(c.Want, got))
		}
	}
}
