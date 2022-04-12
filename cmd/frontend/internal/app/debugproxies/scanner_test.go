package debugproxies

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterScan(t *testing.T) {
	var eps []Endpoint

	consumer := func(seen []Endpoint) {
		eps = nil
		eps = append(eps, seen...)
	}

	// test setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fake.NewSimpleClientset()
	const ns = "test-ns"
	cs := &clusterScanner{
		client:    client.CoreV1(),
		consume:   consumer,
		namespace: ns,
	}
	endpoints := []v1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "gitserver"},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{
					Hostname: "gitserver-0",
					IP:       "192.168.10.0",
				}},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "searcher"},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{
					IP: "192.168.10.3",
				}},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "no-port"},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{
					IP: "192.168.10.1",
				}},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "no-prom-port"},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{
					IP: "192.168.10.2",
				}},
				Ports: []v1.EndpointPort{{
					Port: 2324,
				}},
			}},
		},
	}
	for _, e := range endpoints {
		_, err := cs.client.Endpoints(ns).Create(ctx, &e, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("unable to create test endpoint: %v", err)
		}
	}
	svcs := []v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gitserver",
				Namespace: ns,
				Annotations: map[string]string{
					"sourcegraph.prometheus/scrape": "true",
					"prometheus.io/port":            "2323",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "searcher",
				Annotations: map[string]string{
					"sourcegraph.prometheus/scrape": "true",
					"prometheus.io/port":            "2323",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "no-scrape",
				Annotations: map[string]string{
					"prometheus.io/port": "2323",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "no-prom-port",
				Annotations: map[string]string{
					"sourcegraph.prometheus/scrape": "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "no-port",
				Annotations: map[string]string{
					"sourcegraph.prometheus/scrape": "true",
				},
			},
		},
	}
	for _, svc := range svcs {
		_, err := cs.client.Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}

	cs.scanCluster(ctx)

	want := []Endpoint{{
		Service:  "gitserver",
		Addr:     "192.168.10.0:2323",
		Hostname: "gitserver-0",
	}, {
		Service: "searcher",
		Addr:    "192.168.10.3:2323",
	}, {
		Service: "no-prom-port",
		Addr:    "192.168.10.2:2324",
	}}

	sortOpt := cmpopts.SortSlices(func(x Endpoint, y Endpoint) bool {
		return x.Service < y.Service
	})
	if d := cmp.Diff(want, eps, sortOpt); d != "" {
		t.Errorf("mismatch (-want +got):\n%s", d)
	}
}
