package endpoint

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	m := New("http://test")
	expectEndpoints(t, m, nil, "http://test")

	m = New("http://test-1 http://test-2")
	expectEndpoints(t, m, nil, "http://test-1", "http://test-2")
}

func TestStatic(t *testing.T) {
	m := Static("http://test")
	expectEndpoints(t, m, nil, "http://test")

	m = Static("http://test-1", "http://test-2")
	expectEndpoints(t, m, nil, "http://test-1", "http://test-2")
}

func TestExclude(t *testing.T) {
	endpoints := []string{"http://test-1", "http://test-2", "http://test-3", "http://test-4"}
	m := Static(endpoints...)

	exclude := map[string]bool{}
	for len(endpoints) > 0 {
		expectEndpoints(t, m, exclude, endpoints...)

		exclude[endpoints[len(endpoints)-1]] = true
		endpoints = endpoints[:len(endpoints)-1]
	}
}

func expectEndpoints(t *testing.T, m *Map, exclude map[string]bool, endpoints ...string) {
	t.Helper()

	// We ask for the URL of a large number of keys, we expect to see every
	// endpoint and only those endpoints.
	count := map[string]int{}
	for _, e := range endpoints {
		count[e] = 0
	}
	for i := 0; i < len(endpoints)*10; i++ {
		v, err := m.Get(fmt.Sprintf("test-%d", i), exclude)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if _, ok := count[v]; !ok {
			t.Fatalf("map returned unexpected endpoint %v. Valid: %v", v, endpoints)
		}
		count[v] = count[v] + 1
	}
	t.Logf("counts: %v", count)
	for e, c := range count {
		if c == 0 {
			t.Fatalf("map never returned %v", e)
		}
	}

	// Ensure GetMany matches Get
	var keys, vals []string
	for i := 0; i < len(endpoints)*10; i++ {
		keys = append(keys, fmt.Sprintf("test-%d", i))
		v, err := m.Get(keys[i], nil)
		if err != nil {
			t.Fatalf("Get for GetMany failed: %v", err)
		}
		vals = append(vals, v)
	}
	if got, err := m.GetMany(keys...); err != nil {
		t.Fatalf("GetMany failed: %v", err)
	} else if !reflect.DeepEqual(got, vals) {
		t.Fatalf("GetMany(%v) unexpected response:\ngot  %v\nwant %v", keys, got, vals)
	}
}

func TestEndpoints(t *testing.T) {
	eps := []string{"http://test-1", "http://test-2", "http://test-3", "http://test-4"}
	want := map[string]struct{}{}
	for _, addr := range eps {
		want[addr] = struct{}{}
	}

	m := Static(eps...)
	got, err := m.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Endpoints() unexpected return:\ngot:  %v\nwant: %v", got, want)
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
