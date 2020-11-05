package debugproxies

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

//type k8sTestClient struct {
//	listResponse *corev1.ServiceList
//	getResponses map[string]*corev1.Endpoints
//}
//
//func (ktc *k8sTestClient) Watch(ctx context.Context, namespace string, r k8s.Resource, options ...k8s.Option) (*k8s.Watcher, error) {
//	// we don't use it for tests yet, once we do we need to mock the returned watcher too
//	return nil, errors.New("not implemented")
//}
//
//func (ktc *k8sTestClient) List(ctx context.Context, namespace string, resp k8s.ResourceList, options ...k8s.Option) error {
//	sxs := resp.(*corev1.ServiceList)
//
//	sxs.Items = ktc.listResponse.Items
//	sxs.Metadata = ktc.listResponse.Metadata
//	return nil
//}
//
//func (ktc *k8sTestClient) Get(ctx context.Context, namespace, name string, resp k8s.Resource, options ...k8s.Option) error {
//	ep := ktc.getResponses[name]
//	if ep == nil {
//		return fmt.Errorf("resource with name %s not set up as fixture", name)
//	}
//
//	rep := resp.(*corev1.Endpoints)
//
//	rep.Metadata = ep.Metadata
//	rep.Subsets = ep.Subsets
//	return nil
//}
//
//func (ktc *k8sTestClient) Namespace() string {
//	return "foospace"
//}
//
//func stringPtr(val string) *string {
//	str := val
//	return &str
//}
//
//func int32Ptr(val int32) *int32 {
//	i := val
//	return &i
//}

func TestClusterScan(t *testing.T) {
	var eps []Endpoint

	consumer := func(seen []Endpoint) {
		eps = nil
		eps = append(eps, seen...)
	}

	//ktc := &k8sTestClient{
	//	getResponses: make(map[string]*corev1.Endpoints),
	//}
	client := fake.NewSimpleClientset()
	const ns = "test-ns"

	cs := &clusterScanner{
		client:    client.CoreV1(),
		consume:   consumer,
		namespace: ns,
	}

	// fill up
	var endpoints []v1.Endpoints

	e := v1.Endpoints{
		Subsets: []v1.EndpointSubset{{
			Addresses: []v1.EndpointAddress{{
				Hostname: ("gitserver-0"),
				IP:       ("192.168.10.0"),
			}},
		}},
	}
	endpoints = append(endpoints, e)

	e2 := v1.Endpoints{
		Subsets: []v1.EndpointSubset{{
			Addresses: []v1.EndpointAddress{{
				IP: ("192.168.10.3"),
			}},
		}},
	}
	endpoints = append(endpoints, e2)

	//ktc.getResponses["no-port"] = &v1.Endpoints{
	//	Subsets: []*v1.EndpointSubset{{
	//		Addresses: []*v1.EndpointAddress{{
	//			Ip: stringPtr("192.168.10.1"),
	//		}},
	//	}},
	//}
	//ktc.getResponses["no-prom-port"] = &v1.Endpoints{
	//	Subsets: []*v1.EndpointSubset{{
	//		Addresses: []*v1.EndpointAddress{{
	//			Ip: stringPtr("192.168.10.2"),
	//		}},
	//		Ports: []*v1.EndpointPort{{
	//			Port: int32Ptr(2324),
	//		}},
	//	}},
	//}

	for _, e := range endpoints {
		_, err := client.CoreV1().Endpoints(ns).Create(&e)
		if err != nil {
			t.Fatalf("unable to create test endpoint")
		}
	}
	//metav1&metav1.ObjectMeta{
	//	Namespace: ("foospace"),
	//	Name:      ("gitserver"),
	//	Annotations: map[string]string{
	//		"sourcegraph.prometheus/scrape": "true",
	//		"prometheus.io/port":            "2323",

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitserver",
			Namespace: ns,
			Annotations: map[string]string{
				"sourcegraph.prometheus/scrape": "true",
				"prometheus.io/port":            "2323",
			},
		},
	}
	//	{
	//		Metadata: &metav1.ObjectMeta{
	//			Namespace: stringPtr("foospace"),
	//			Name:      stringPtr("searcher"),
	//			Annotations: map[string]string{
	//				"sourcegraph.prometheus/scrape": "true",
	//				"prometheus.io/port":            "2323",
	//			},
	//		},
	//	},
	//	{
	//		Metadata: &metav1.ObjectMeta{
	//			Namespace: stringPtr("foospace"),
	//			Name:      stringPtr("no-scrape"),
	//			Annotations: map[string]string{
	//				"prometheus.io/port": "2323",
	//			},
	//		},
	//	},
	//	{
	//		Metadata: &metav1.ObjectMeta{
	//			Namespace: stringPtr("foospace"),
	//			Name:      stringPtr("no-prom-port"),
	//			Annotations: map[string]string{
	//				"sourcegraph.prometheus/scrape": "true",
	//			},
	//		},
	//	},
	//	{
	//		Metadata: &metav1.ObjectMeta{
	//			Namespace: stringPtr("foospace"),
	//			Name:      stringPtr("no-port"),
	//			Annotations: map[string]string{
	//				"sourcegraph.prometheus/scrape": "true",
	//			},
	//		},
	//	},
	//},

	_, err := client.CoreV1().Services(ns).Create(svc)
	if err != nil {
		t.Fatal(err)
	}

	cs.scanCluster()

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

	if !cmp.Equal(want, eps) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, eps))
	}
}
