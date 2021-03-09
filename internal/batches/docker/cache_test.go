package docker

import "testing"

func TestImageCache(t *testing.T) {
	cache := NewImageCache()
	if cache == nil {
		t.Error("unexpected nil cache")
	}

	have := cache.Get("foo")
	if have == nil {
		t.Error("unexpected nil error")
	}
	if name := have.(*image).name; name != "foo" {
		t.Errorf("invalid name: have=%q want=%q", name, "foo")
	}

	again := cache.Get("foo")
	if have != again {
		t.Errorf("invalid memoisation: first=%v second=%v", have, again)
	}
}
