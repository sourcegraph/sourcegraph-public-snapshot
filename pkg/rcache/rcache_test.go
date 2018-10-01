package rcache

import "testing"

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
		t.Fatal("Initial Get should of found nothing")
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
