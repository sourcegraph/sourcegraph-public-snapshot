package endpoint

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Tests that even with publishNotReadyAddresses: true in the headless services,
// the endpoints are not always returned during rollouts. Just start a roll-out in
// dogfood and run this test.
//
// k -n dogfood-k8s rollout restart sts/indexed-search
// go test -integration -run=Integration
// === RUN   TestIntegrationK8SNotReadyAddressesBug
// DBUG[08-11|16:49:14] kubernetes endpoints  service=indexed-search urls="[indexed-search-0.indexed-search indexed-search-1.indexed-search]"
// DBUG[08-11|16:49:14] kubernetes endpoints  service=indexed-search urls="[indexed-search-0.indexed-search indexed-search-1.indexed-search]"
// DBUG[08-11|16:49:28] kubernetes endpoints  service=indexed-search urls="[indexed-search-0.indexed-search indexed-search-1.indexed-search]"
// DBUG[08-11|16:49:40] kubernetes endpoints  service=indexed-search urls=[indexed-search-0.indexed-search]
//     endpoint_test.go:163: endpoint set has shrunk from 2 to 1
// 	--- FAIL: TestIntegrationK8SNotReadyAddressesBug (26.94s)

var integration = flag.Bool("integration", false, "Run integration tests")

func TestIntegrationK8SNotReadyAddressesBug(t *testing.T) {
	if !*integration {
		t.Skip("Not running integration tests")
	}

	urlspec := "k8s+rpc://indexed-search"
	m := Map{
		urlspec:   urlspec,
		discofunk: k8sDiscovery(urlspec, "dogfood-k8s", localClient),
	}

	began := time.Now()
	count := 0
	for time.Since(began) <= time.Minute {
		eps, err := m.Endpoints()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("endpoints: %v", eps)

		if count == 0 {
			count = len(eps)
		} else if len(eps) < count {
			t.Fatalf("endpoint set shrunk from %d to %d", count, len(eps))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func TestIntegrationK8SStatefulSet(t *testing.T) {
	if !*integration {
		t.Skip("Not running integration tests")
	}

	urlspec := "k8s+rpc://indexed-search?kind=sts"
	m := Map{
		urlspec:   urlspec,
		discofunk: k8sDiscovery(urlspec, "dogfood-k8s", localClient),
	}

	t.Log(m.Endpoints())
}

func TestK8sURL(t *testing.T) {
	endpoint := "endpoint.service"
	cases := map[string]string{
		"k8s+http://searcher:3181":          "http://endpoint.service:3181",
		"k8s+http://searcher":               "http://endpoint.service",
		"k8s+http://searcher.namespace:123": "http://endpoint.service:123",
		"k8s+rpc://indexed-search:6070":     "endpoint.service:6070",
	}
	for rawurl, want := range cases {
		u, err := parseURL(rawurl)
		if err != nil {
			t.Fatal(err)
		}
		got := u.endpointURL(endpoint)
		if got != want {
			t.Errorf("mismatch on %s (-want +got):\n%s", rawurl, cmp.Diff(want, got))
		}
	}
}

func localClient() (*kubernetes.Clientset, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
