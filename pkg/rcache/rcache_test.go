package rcache

import "testing"

func clearAll(t *testing.T, prefix string) {
	if err := ClearAllForTest(prefix); err != nil {
		t.Fatal(err)
	}
}

func TestCache_namespace(t *testing.T) {
	globalPrefix = "__test__TestCache_namespace"
	clearAll(t, globalPrefix)
	defer clearAll(t, globalPrefix)

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
		caches[i] = New(test.prefix, 123)
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
	globalPrefix = "__test__TestCache_simple"
	clearAll(t, globalPrefix)
	defer clearAll(t, globalPrefix)

	c := New("some_prefix", 123)
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
