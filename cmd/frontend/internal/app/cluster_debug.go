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

type proxyEndpoint struct {
	reverseProxy http.Handler
	service      string
	displayName  string
	host         string
}

func (pe *proxyEndpoint) String() string {
	var sb strings.Builder

	sb.WriteString(pe.service)
	sb.WriteString("-")
	sb.WriteString(pe.host)

	return sb.String()
}

type clusterInstrumenter struct {
	sync.RWMutex
	reverseProxies map[string]*proxyEndpoint
	client         *k8s.Client
	once           sync.Once
}

func (ci *clusterInstrumenter) runEventLoop() {
	for {
		err := ci.watchEndpointEvents()
		log15.Debug("failed to watch kubernetes endpoints", "error", err)
		time.Sleep(time.Second * 5)
	}
}

type hostInfo struct {
	service     string
	displayName string
}

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

func proxyEndpointFromHost(host string, pathPrefix string) http.Handler {
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

func (ci *clusterInstrumenter) populate(hostsToServices map[string]hostInfo) {
	rps := make(map[string]*proxyEndpoint, len(hostsToServices))
	for host, hi := range hostsToServices {
		rps[hi.displayName] = &proxyEndpoint{
			reverseProxy: proxyEndpointFromHost(host, "/-/debug/cluster/"+hi.displayName),
			service:      hi.service,
			displayName:  hi.displayName,
			host:         host,
		}
	}

	ci.Lock()
	ci.reverseProxies = rps
	ci.Unlock()
}

func addrToHost(addr *corev1.EndpointAddress, port int) string {
	if addr.Ip != nil {
		return fmt.Sprintf("%s:%d", *addr.Ip, port)
	} else if addr.Hostname != nil && *addr.Hostname != "" {
		return fmt.Sprintf("%s:%d", *addr.Hostname, port)
	}
	return ""
}

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

func newClusterInstrumenter(client *k8s.Client) *clusterInstrumenter {
	ci := &clusterInstrumenter{
		client: client,
	}
	ci.scanCluster()
	go ci.runEventLoop()
	return ci
}
