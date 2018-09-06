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

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

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

		endpoints, err := cl.CoreV1().Endpoints(u.Namespace).List(meta_v1.ListOptions{FieldSelector: "metadata.name=" + u.Service})
		if err != nil {
			return nil, err
		}
		b := &urlMapBuilder{K8sURL: u}
		for _, ep := range endpoints.Items {
			b.Add(&ep)
		}

		// Kick off watcher in the background
		go inform(cl, m, u)

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

	lw := cache.NewListWatchFromClient(cl.CoreV1().RESTClient(), "endpoints", u.Namespace, fields.ParseSelectorOrDie("metadata.name="+u.Service))
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
		b := &urlMapBuilder{K8sURL: u}
		store := inf.GetStore()
		for _, o := range store.List() {
			b.Add(o.(*v1.Endpoints))
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
func (b *urlMapBuilder) Add(ep *v1.Endpoints) {
	if ep.ObjectMeta.Name != b.K8sURL.Service {
		return
	}
	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			b.urls = append(b.urls, b.K8sURL.endpointURL(addr.IP))
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
