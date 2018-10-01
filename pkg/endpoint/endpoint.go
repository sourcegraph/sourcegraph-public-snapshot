// Package endpoint provides a consistent hash map for URLs to kubernetes
// endpoints.
package endpoint

import (
	"context"
	"fmt"
	"hash/crc32"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
)

// Map is a consistent hash map to URLs. It uses the kubernetes API to watch
// the endpoints for a service and update the map when they change. It can
// also fallback to static URLs if not configured for kubernetes.
type Map struct {
	mu   sync.Mutex
	init func() (*hashMap, error)
	err  error
	urls *hashMap
}

// New creates a new Map for rawurl. We treat schemes prefixed with k8s+
// specially. The expected format of that is
// k8s+http://service.namespace:port/path. namespace, port and path is
// optional. URLs of this form will consistently hash amongst the endpoints
// for the service. The values returned by Get will look like
// http://endpoint:port/path.
//
// Example: rawurl is k8s+http://searcher
func New(rawurl string) *Map {
	if !strings.HasPrefix(rawurl, "k8s+") {
		// Non-k8s urls we return a static map
		return &Map{urls: newConsistentHashMap([]string{rawurl})}
	}

	m := &Map{}

	// Kick off setting the initial urls or err on first access. We don't rely
	// just on inform since it may not communicate updates.
	m.init = func() (*hashMap, error) {
		u, err := parseURL(rawurl)
		if err != nil {
			return nil, err
		}

		client, err := loadClient()
		if err != nil {
			return nil, err
		}

		var endpoints corev1.Endpoints
		err = client.Get(context.Background(), client.Namespace, u.Service, &endpoints)
		if err != nil {
			return nil, err
		}
		b := &urlMapBuilder{K8sURL: u}
		b.Add(&endpoints)

		// Kick off watcher in the background
		go func() {
			for {
				err := inform(client, m, u)
				log15.Debug("failed to watch kubernetes endpoint", "name", u.Service, "error", err)
				time.Sleep(time.Second)
			}
		}()

		return b.Build()
	}

	return m
}

// Get the closest URL in the hash to the provided key that is not in
// exclude. If no URL is found, "" is returned.
//
// Note: For k8s URLs we return URLs based on the registered endpoints. The
// endpoint may not actually be available yet / at the moment. So users of the
// URL should implement a retry strategy.
func (m *Map) Get(key string, exclude map[string]bool) (string, error) {
	m.mu.Lock()
	if m.init != nil {
		m.urls, m.err = m.init()
		m.init = nil // prevent running again
	}
	urls, err := m.urls, m.err
	m.mu.Unlock()

	if err != nil {
		return "", err
	}
	return urls.get(key, exclude), nil
}

func inform(client *k8s.Client, m *Map, u *k8sURL) error {
	watcher, err := client.Watch(context.Background(), client.Namespace, new(corev1.Endpoints), k8s.QueryParam("fieldSelector", "metadata.name="+u.Service))
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		var endpoints corev1.Endpoints
		eventType, err := watcher.Next(&endpoints)
		if err != nil {
			return err
		}

		b := &urlMapBuilder{K8sURL: u}
		if eventType == k8s.EventAdded || eventType == k8s.EventModified {
			b.Add(&endpoints)
		}
		urls, err := b.Build()
		m.mu.Lock()
		m.urls, m.err = urls, err
		m.mu.Unlock()
	}
}

type urlMapBuilder struct {
	K8sURL *k8sURL
	urls   []string
}

// Add adds all addresses associated with the endpoint for the Service.
func (b *urlMapBuilder) Add(ep *corev1.Endpoints) {
	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.Ip != nil {
				b.urls = append(b.urls, b.K8sURL.endpointURL(*addr.Ip))
			}
		}
	}
}

func (b *urlMapBuilder) Build() (*hashMap, error) {
	if len(b.urls) == 0 {
		return nil, errors.Errorf("No %s endpoints", b.K8sURL.Service)
	}
	return newConsistentHashMap(b.urls), nil
}

type k8sURL struct {
	url.URL

	Service   string
	Namespace string
}

func (u *k8sURL) endpointURL(endpoint string) string {
	uCopy := u.URL
	if port := u.Port(); port != "" {
		uCopy.Host = endpoint + ":" + port
	} else {
		uCopy.Host = endpoint
	}
	return uCopy.String()
}

func parseURL(rawurl string) (*k8sURL, error) {
	u, err := url.Parse(strings.TrimPrefix(rawurl, "k8s+"))
	if err != nil {
		return nil, err
	}
	parts := strings.Split(u.Hostname(), ".")
	var svc, ns string
	switch len(parts) {
	case 1:
		svc = parts[0]
	case 2:
		svc, ns = parts[1], parts[2]
	default:
		return nil, fmt.Errorf("invalid k8s url. expected k8s+http://service.namespace:port/path, got %s", rawurl)
	}
	return &k8sURL{
		URL:       *u,
		Service:   svc,
		Namespace: ns,
	}, nil
}

func newConsistentHashMap(keys []string) *hashMap {
	// 50 replicas and crc32.ChecksumIEEE are the defaults used by
	// groupcache.
	m := hashMapNew(50, crc32.ChecksumIEEE)
	m.add(keys...)
	return m
}

func loadClient() (*k8s.Client, error) {
	// Uncomment below to test against a real cluster. This is only important
	// when you are changing how we interact with the k8s API and you want to
	// test against the real thing. Remember to run a proxy to the API first
	// (takes care of auth and endpoint details)
	//
	//   kubectl proxy --port 10810
	//
	// NewInClusterClient only works when running inside of a pod in a k8s
	// cluster.
	/*
		return &k8s.Client{
			Endpoint:  "http://127.0.0.1:10810",
			Namespace: "prod",
			Client:    http.DefaultClient,
			//Client: &http.Client{Transport: &loghttp.Transport{}},
		}, nil
	*/
	return k8s.NewInClusterClient()
}
