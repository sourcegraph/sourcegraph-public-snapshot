// Package endpoint provides a consistent hash map for URLs to kubernetes
// endpoints.
package endpoint

import (
	"context"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/groupcache/consistenthash"
	"github.com/pkg/errors"
)

// Map is a consistent hash map to URLs. It uses the kubernetes API to watch
// the endpoints for a service and update the map when they change. It can
// also fallback to static URLs if not configured for kubernetes.
type Map struct {
	mu   sync.Mutex
	urls *consistenthash.Map
}

// New creates a new Map for rawurl. We treat schemes prefixed with k8s+
// specially. The expected format of that is
// k8s+http://service.namespace:port/path. namespace, port and path is
// optional. URLs of this form will consistently hash amongst the endpoints
// for the service. The values returned by Get will look like
// http://endpoint:port/path.
//
// Example: rawurl is k8s+http://searcher
func New(rawurl string) (*Map, error) {
	if !strings.HasPrefix(rawurl, "k8s+") {
		// Non-k8s urls we return a static map
		return &Map{
			urls: newConsistentHashMap([]string{rawurl}),
		}, nil
	}

	u, err := parseURL(rawurl)
	if err != nil {
		return nil, err
	}

	cl, err := client()
	if err != nil {
		return nil, err
	}

	if u.Namespace == "" {
		u.Namespace = podNamespace()
	}
	if u.Namespace == "" {
		return nil, errors.Errorf("%s does not specify namespace and could not detect pod namespace", rawurl)
	}

	// Initially we return a map which just returns the URL for the
	// service. Once we start watching the endpoints it will be updated.
	m := &Map{
		urls: newConsistentHashMap([]string{u.serviceURL()}),
	}

	// Kick off watcher in the background
	go inform(cl, m, u)

	return m, nil
}

// Get the closest URL in the hash to the provided key.
func (m *Map) Get(key string) string {
	m.mu.Lock()
	u := m.urls.Get(key)
	m.mu.Unlock()
	return u
}

func inform(cl *kubernetes.Clientset, m *Map, u *k8sURL) {
	// We ignore the update events, and instead use them as a signal to
	// re-read what is in the informer. A message available to read on
	// updateC means we should recheck the informer.
	updateC := make(chan struct{}, 1)
	shouldCheckStore := func() {
		select {
		case updateC <- struct{}{}:
		default:
		}
	}

	lw := cache.NewListWatchFromClient(cl.Core().RESTClient(), "endpoints", u.Namespace, fields.ParseSelectorOrDie("metadata.name="+u.Service))
	inf := cache.NewSharedInformer(lw, &v1.Endpoints{}, 5*time.Minute)
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(o interface{}) {
			shouldCheckStore()
		},
		UpdateFunc: func(_, o interface{}) {
			shouldCheckStore()
		},
		DeleteFunc: func(o interface{}) {
			shouldCheckStore()
		},
	})

	go inf.Run(context.Background().Done())

	for {
		<-updateC
		var urls []string
		store := inf.GetStore()
		for _, o := range store.List() {
			ep := o.(*v1.Endpoints)
			if ep.ObjectMeta.Name != u.Service {
				continue
			}
			for _, subset := range ep.Subsets {
				for _, addr := range subset.Addresses {
					urls = append(urls, u.endpointURL(addr.IP))
				}
			}
		}
		if len(urls) == 0 {
			urls = []string{u.serviceURL()}
		}
		urlMap := newConsistentHashMap(urls)
		m.mu.Lock()
		m.urls = urlMap
		m.mu.Unlock()
	}
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

func (u *k8sURL) serviceURL() string {
	return u.URL.String()
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

func newConsistentHashMap(keys []string) *consistenthash.Map {
	// 50 replicas and crc32.ChecksumIEEE are the defaults used by
	// groupcache.
	m := consistenthash.New(50, crc32.ChecksumIEEE)
	m.Add(keys...)
	return m
}

func client() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

// podNamespace returns the namespace for a pod. It is based on
// k8s.io/client-go/tools/clientcmd.inClusterClientConfig.Namespace
func podNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}
