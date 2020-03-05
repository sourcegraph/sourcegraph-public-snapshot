package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"gopkg.in/inconshreveable/log15.v2"
)

// proxyEndpoint couples the reverse proxy with the endpoint it proxies.
type proxyEndpoint struct {
	reverseProxy http.Handler
	service      string
	// displayName is unique within the endpoints and is suitable as a key
	displayName string
	host        string
}

func (pe *proxyEndpoint) String() string {
	return fmt.Sprintf("%s-%s", pe.service, pe.host)
}

// clusterInstrumenter holds the state to proxy endpoints in a cluster
// and it handles serving the index page and routing the requests being proxied to their
// respective reverse proxy. It knows how to scan a cluster for endpoints meeting the scraping criteria
// (sourcegraph.prometheus/scrape=true) and knows how to keep that info uptodate.
type clusterInstrumenter struct {
	// protects the reverseProxies map
	sync.RWMutex
	// keys are the displayNames
	reverseProxies map[string]*proxyEndpoint
	client         *k8s.Client
}

// Runs the k8s.Watch endpoints event loop, and it triggers a rescan of cluster when something changes with endpoints.
func (ci *clusterInstrumenter) runEventLoop() {
	for {
		err := ci.watchEndpointEvents()
		log15.Debug("failed to watch kubernetes endpoints", "error", err)
		time.Sleep(time.Second * 5)
	}
}

// hostInfo couples a service with the displayName of one endpoint
type hostInfo struct {
	service     string
	displayName string
}

// scanCluster looks for endpoints belonging to services that have annotation sourcegraph.prometheus/scrape=true.
// It derives the appropriate port from the prometheus.io/port annotation.
func (ci *clusterInstrumenter) scanCluster() {
	var services corev1.ServiceList

	err := ci.client.List(context.Background(), ci.client.Namespace, &services)
	if err != nil {
		log15.Error("k8s failed to list services", "error", err)
		return
	}

	hostsToServices := make(map[string]hostInfo)

	for _, svc := range services.Items {
		svcName := *svc.Metadata.Name

		// TODO(uwedeportivo): pgsql doesn't work, figure out why
		if svcName == "pgsql" {
			continue
		}

		if svc.Metadata.Annotations["sourcegraph.prometheus/scrape"] != "true" {
			continue
		}

		portStr := svc.Metadata.Annotations["prometheus.io/port"]
		if portStr == "" {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log15.Debug("k8s prometheus.io/port annotation for service is not an integer", "service", svcName, "port", portStr)
			continue
		}

		var endpoints corev1.Endpoints
		err = ci.client.Get(context.Background(), ci.client.Namespace, svcName, &endpoints)
		if err != nil {
			log15.Error("k8s failed to get endpoints", "error", err)
			return
		}

		for _, subset := range endpoints.Subsets {
			for i, addr := range subset.Addresses {
				host := addrToHost(addr, port)
				if host != "" {
					hostsToServices[host] = hostInfo{
						service:     svcName,
						displayName: fmt.Sprintf("%s-%d", svcName, i),
					}
				}
			}
		}
	}

	ci.populate(hostsToServices)
}

// reverseProxyFromHost creates a reverse proxy from specified host with the path prefix that will be stripped from
// request before it gets sent to the destination endpoint.
func reverseProxyFromHost(host string, pathPrefix string) http.Handler {
	return adminOnly(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = host
			if i := strings.Index(req.URL.Path, pathPrefix); i >= 0 {
				req.URL.Path = req.URL.Path[i+len(pathPrefix):]
			}
		},
		ErrorLog: log.New(env.DebugOut, fmt.Sprintf("k8s %s debug proxy: ", host), log.LstdFlags),
	})
}

// populate updates the clusterIntrumenter state with the scan results.
// goroutine-safe
func (ci *clusterInstrumenter) populate(hostsToServices map[string]hostInfo) {
	rps := make(map[string]*proxyEndpoint, len(hostsToServices))
	for host, hi := range hostsToServices {
		rps[hi.displayName] = &proxyEndpoint{
			reverseProxy: reverseProxyFromHost(host, "/-/debug/cluster/"+hi.displayName),
			service:      hi.service,
			displayName:  hi.displayName,
			host:         host,
		}
	}

	ci.Lock()
	ci.reverseProxies = rps
	ci.Unlock()
}

// addrToHost converts a scanned k8s endpoint address structure into a string that is the host:port part of a URL.
func addrToHost(addr *corev1.EndpointAddress, port int) string {
	if addr.Ip != nil {
		return fmt.Sprintf("%s:%d", *addr.Ip, port)
	} else if addr.Hostname != nil && *addr.Hostname != "" {
		return fmt.Sprintf("%s:%d", *addr.Hostname, port)
	}
	return ""
}

// watchEndpointEvents use the k8s watch API operation to watch for endpoint events. Spins forever unless an error
// occurs that would necessitate creating a new watcher. The caller will then call again creating the new watcher.
func (ci *clusterInstrumenter) watchEndpointEvents() error {
	watcher, err := ci.client.Watch(context.Background(), ci.client.Namespace, new(corev1.Endpoints))
	if err != nil {
		return fmt.Errorf("k8s client.Watch error: %w", err)
	}
	defer watcher.Close()

	for {
		var eps corev1.Endpoints
		eventType, err := watcher.Next(&eps)
		if err != nil {
			// we need a new watcher
			return fmt.Errorf("k8s watcher.Next error: %w", err)
		}

		if eventType == k8s.EventError {
			// we need a new watcher
			return errors.New("error event")
		}

		ci.scanCluster()
	}
}

// ServeIndex composes the simple index page with the endpoints sorted by their displayName.
func (ci *clusterInstrumenter) ServeIndex(w http.ResponseWriter, r *http.Request) {
	ci.RLock()
	displayNames := make([]string, 0, len(ci.reverseProxies))
	for displayName := range ci.reverseProxies {
		displayNames = append(displayNames, displayName)
	}
	ci.RUnlock()

	sort.Strings(displayNames)

	for _, displayName := range displayNames {
		fmt.Fprintf(w, `<a href="cluster/%s/">%s</a><br>`, displayName, displayName)
	}
}

// ServeReverseProxy routes the request to the appropriate reverse proxy by splitting the request path and finding
// the displayName.
func (ci *clusterInstrumenter) ServeReverseProxy(w http.ResponseWriter, r *http.Request) error {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("k8s service missing"),
		}
	}

	ci.RLock()
	pe := ci.reverseProxies[pathParts[4]]
	ci.RUnlock()

	if pe == nil {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("k8s endpoint disappeared from cluster"),
		}
	}

	pe.reverseProxy.ServeHTTP(w, r)
	return nil
}

// newClusterInstrumenter creates the new cluster instrumenter, kicks of an initial cluster scan and starts up the
// goroutine that watches for endpoint events.
func newClusterInstrumenter(client *k8s.Client) *clusterInstrumenter {
	ci := &clusterInstrumenter{
		client: client,
	}
	ci.scanCluster()
	go ci.runEventLoop()
	return ci
}
