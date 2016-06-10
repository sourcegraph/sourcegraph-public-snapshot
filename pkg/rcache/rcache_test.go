package rcache

import (
	"reflect"
	"testing"
)

func clearAll(t *testing.T, prefix string) {
	if err := ClearAllForTest(prefix); err != nil {
		t.Fatal(err)
	}
}

func TestRedis(t *testing.T) {
	clearAll(t, globalPrefix)
	globalPrefix = "__test__TestRedis"
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

	caches := make([]*Redis, len(cases))
	for i, test := range cases {
		caches[i] = New(test.prefix)
		for _, entry := range test.entries {
			if err := caches[i].Add(entry.k, entry.v, entry.ttl); err != nil {
				t.Fatalf("error adding entry to redis (prefix=%s): %s", test.prefix, err)
			}
		}
	}
	for i, test := range cases {
		// test all the keys that should be present are found
		for _, entry := range test.entries {
			var res string
			if err := caches[i].Get(entry.k, &res); err != nil {
				t.Fatalf("error getting entry from redis (prefix=%s): %s", test.prefix, err)
			}
			if res != entry.v {
				t.Errorf("expected %s, got %s", entry.v, res)
			}
		}

		// test not found case
		var res interface{}
		err := caches[i].Get("not-found", &res)
		if err != ErrNotFound {
			t.Errorf("expected error %q, got %q", err, ErrNotFound)
		}
	}
}

func TestRedis_values(t *testing.T) {
	clearAll(t, globalPrefix)
	globalPrefix = "__test__TestRedis_values"
	defer clearAll(t, globalPrefix)

	cache := New("some_prefix")

	{
		var v1, v1_got string
		v1 = "asdf"
		if err := cache.Add("k1", v1, -1); err != nil {
			t.Fatal(err)
		}
		if err := cache.Get("k1", &v1_got); err != nil {
			t.Fatal(err)
		}
		if v1 != v1_got {
			t.Errorf("expected %s, got %s", v1, v1_got)
		}
	}

	{
		var v1, v1_got map[string]string
		v1 = map[string]string{"a": "1", "b": "2", "c": "3"}
		if err := cache.Add("k1", v1, -1); err != nil {
			t.Fatal(err)
		}
		if err := cache.Get("k1", &v1_got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v1, v1_got) {
			t.Errorf("expected %s, got %s", v1, v1_got)
		}
	}

	{
		var v1, v1_got []string
		v1 = []string{"1", "2", "3"}
		if err := cache.Add("k1", v1, -1); err != nil {
			t.Fatal(err)
		}
		if err := cache.Get("k1", &v1_got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v1, v1_got) {
			t.Errorf("expected %s, got %s", v1, v1_got)
		}
	}

	{
		type X struct {
			A string
			B int
		}
		var v1, v1_got X
		v1 = X{A: "asdf", B: 23}
		if err := cache.Add("k1", v1, -1); err != nil {
			t.Fatal(err)
		}
		if err := cache.Get("k1", &v1_got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v1, v1_got) {
			t.Errorf("expected %v, got %v", v1, v1_got)
		}
	}
}
