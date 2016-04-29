package cache

import (
	"testing"

	"github.com/golang/groupcache/lru"
)

func TestHook(t *testing.T) {
	var status string
	c := Hook(lru.New(5), func() { status = "hit" }, func() { status = "miss" })

	status = ""
	_, ok := c.Get("a")
	if ok || status != "miss" {
		t.Errorf("Expected miss: ok=%v status=%v", ok, status)
	}

	c.Add("a", "b")

	status = ""
	_, ok = c.Get("a")
	if !ok || status != "hit" {
		t.Errorf("Expected hit: ok=%v status=%v", ok, status)
	}
}
