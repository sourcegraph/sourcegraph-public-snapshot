package endpoint

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	m := New("http://test")
	expectEndpoints(t, m, "http://test")

	m = New("http://test-1 http://test-2")
	expectEndpoints(t, m, "http://test-1", "http://test-2")
}

func TestStatic(t *testing.T) {
	m := Static("http://test")
	expectEndpoints(t, m, "http://test")

	m = Static("http://test-1", "http://test-2")
	expectEndpoints(t, m, "http://test-1", "http://test-2")
}

func TestGetN(t *testing.T) {
	endpoints := []string{"http://test-1", "http://test-2", "http://test-3", "http://test-4"}
	m := Static(endpoints...)

	node, _ := m.Get("foo")
	have, _ := m.GetN("foo", 3)

	if len(have) != 3 {
		t.Fatalf("GetN(3) didn't return 3 nodes")
	}

	if have[0] != node {
		t.Fatalf("GetN(foo, 3)[0] != Get(foo): %s != %s", have[0], node)
	}

	want := []string{"http://test-4", "http://test-2", "http://test-1"}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("GetN(\"foo\", 3):\nhave: %v\nwant: %v", have, want)
	}
}

func expectEndpoints(t *testing.T, m *Map, endpoints ...string) {
	t.Helper()

	// We ask for the URL of a large number of keys, we expect to see every
	// endpoint and only those endpoints.
	count := map[string]int{}
	for _, e := range endpoints {
		count[e] = 0
	}
	for i := 0; i < len(endpoints)*10; i++ {
		v, err := m.Get(fmt.Sprintf("test-%d", i))
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
		v, err := m.Get(keys[i])
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
