package rcache

import (
	"reflect"
	"testing"
)

func TestCache_namespace(t *testing.T) {
	SetupForTest(t)

	type kvTTL struct {
		k   string
		v   string
		ttl int
	}
	type testcase struct {
		prefix  string
		entries []kvTTL
	}

	cases := []testcase{{
		prefix: "a",
		entries: []kvTTL{
			{k: "k0", v: "v0", ttl: -1},
			{k: "k1", v: "v1", ttl: 123},
			{k: "k2", v: "v2", ttl: 456},
		}}, {
		prefix: "b",
		entries: []kvTTL{
			{k: "k0", v: "v0", ttl: 234},
			{k: "k1", v: "v1", ttl: -1},
			{k: "k2", v: "v2", ttl: -1},
		}}, {
		prefix: "c",
		entries: []kvTTL{
			{k: "k0", v: "v0", ttl: -1},
			{k: "k1", v: "v1", ttl: 123},
			{k: "k2", v: "v2", ttl: -1},
		}},
	}

	caches := make([]*Cache, len(cases))
	for i, test := range cases {
		caches[i] = New(test.prefix)
		for _, entry := range test.entries {
			caches[i].Set(entry.k, []byte(entry.v))
		}
	}
	for i, test := range cases {
		// test all the keys that should be present are found
		for _, entry := range test.entries {
			b, ok := caches[i].Get(entry.k)
			if !ok {
				t.Fatalf("error getting entry from redis (prefix=%s)", test.prefix)
			}
			if string(b) != entry.v {
				t.Errorf("expected %s, got %s", entry.v, string(b))
			}
		}

		// test not found case
		if _, ok := caches[i].Get("not-found"); ok {
			t.Errorf("expected not found")
		}
	}
}

func TestCache_simple(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	_, ok := c.Get("a")
	if ok {
		t.Fatal("Initial Get should find nothing")
	}

	c.Set("a", []byte("b"))
	b, ok := c.Get("a")
	if !ok {
		t.Fatal("Expect to get a after setting")
	}
	if string(b) != "b" {
		t.Fatalf("got %v, want %v", string(b), "b")
	}

	c.Delete("a")
	_, ok = c.Get("a")
	if ok {
		t.Fatal("Get after delete should of found nothing")
	}
}

func TestCache_multi(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	vals := c.GetMulti("k0", "k1", "k2")
	if got, exp := vals, [][]byte{nil, nil, nil}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v on initial fetch, got %v", exp, got)
	}

	c.Set("k0", []byte("b"))
	if got, exp := c.GetMulti("k0"), bytes("b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "a"})
	if got, exp := c.GetMulti("k0"), bytes("a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "a"}, [2]string{"k1", "b"})
	if got, exp := c.GetMulti("k0"), bytes("a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k1"), bytes("b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k0", "k1"), bytes("a", "b"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	if got, exp := c.GetMulti("k1", "k0"), bytes("b", "a"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	c.SetMulti([2]string{"k0", "x"}, [2]string{"k1", "y"}, [2]string{"k2", "z"})
	if got, exp := c.GetMulti("k0", "k1", "k2"), bytes("x", "y", "z"); !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
	got, exist := c.Get("k0")
	if exp := "x"; !exist || string(got) != exp {
		t.Errorf("Expected %v, but got %v", exp, string(got))
	}

	c.Delete("k0")
	if got, exp := c.GetMulti("k0", "k1", "k2"), [][]byte{nil, []byte("y"), []byte("z")}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}

func bytes(s ...string) [][]byte {
	if s == nil {
		return nil
	}
	t := make([][]byte, len(s))
	for i, v := range s {
		t[i] = []byte(v)
	}
	return t
}
