package endpoint

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
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
	logger := logtest.Scoped(t)
	if !*integration {
		t.Skip("Not running integration tests")
	}

	urlspec := "k8s+rpc://indexed-search"
	m := Map{
		urlspec:   urlspec,
		discofunk: k8sDiscovery(logger, urlspec, "dogfood-k8s", localClient),
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

func TestIntegrationK8SStatefulSetEquivalence(t *testing.T) {
	if !*integration {
		t.Skip("Not running integration tests")
	}
	logger := logtest.Scoped(t)

	u1 := "k8s+rpc://indexed-search:6070?kind=sts"
	m1 := Map{
		urlspec:   u1,
		discofunk: k8sDiscovery(logger, u1, "prod", localClient),
	}

	u2 := "k8s+rpc://indexed-search:6070"
	m2 := Map{
		urlspec:   u2,
		discofunk: k8sDiscovery(logger, u2, "prod", localClient),
	}

	have, _ := m1.Endpoints()
	want, _ := m2.Endpoints()

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("mismatch (-have, +want):\n%s", diff)
	}
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

func TestK8sEndpoints(t *testing.T) {
	cases := []struct {
		name string
		spec string
		obj  any
		want []string
	}{{
		name: "endpoint empty",
		spec: "k8s+http://searcher:3138",
		obj:  &corev1.Endpoints{},
		want: []string{},
	}, {
		name: "endpoint ip",
		spec: "k8s+http://searcher:3138",
		obj: &corev1.Endpoints{
			Subsets: []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{
					IP: "10.164.38.109",
				}, {
					IP: "10.164.38.110",
				}},
			}},
		},
		want: []string{"http://10.164.38.109:3138", "http://10.164.38.110:3138"},
	}, {
		name: "endpoint hostname",
		spec: "k8s+rpc://indexed-search:6070",
		obj: &corev1.Endpoints{
			Subsets: []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{
					Hostname: "indexed-search-2",
					IP:       "10.164.0.31",
				}, {
					Hostname: "indexed-search-0",
					IP:       "10.164.38.110",
				}},
			}},
		},
		want: []string{"indexed-search-2.indexed-search:6070", "indexed-search-0.indexed-search:6070"},
	}, {
		name: "sts",
		spec: "k8s+rpc://indexed-search:6070?kind=sts",
		obj: &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "indexed-search",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas:    pointers.Ptr(int32(2)),
				ServiceName: "indexed-search-svc", // normally same as sts name, but testing when different
			},
		},
		want: []string{"indexed-search-0.indexed-search-svc:6070", "indexed-search-1.indexed-search-svc:6070"},
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u, err := parseURL(c.spec)
			if err != nil {
				t.Fatal(err)
			}

			got := k8sEndpoints(u, c.obj)
			if d := cmp.Diff(c.want, got, cmpopts.EquateEmpty()); d != "" {
				t.Fatalf("unexpected endpoints (-want, +got):\n%s", d)
			}
		})
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
