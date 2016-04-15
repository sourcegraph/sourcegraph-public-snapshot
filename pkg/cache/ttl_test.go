package cache

import (
	"testing"
	"time"

	"github.com/golang/groupcache/lru"
)

func TestTTL(t *testing.T) {
	now := time.Now()
	ttl := 100 * time.Millisecond
	c := TTL(lru.New(5), ttl)
	c.Add("foo", "bar")
	if v, found := c.Get("foo"); !found || v.(string) != "bar" {
		if time.Since(now) >= ttl {
			t.Skip("Machine is too resource constrained to run test")
			return
		}
		t.Error("TTL cache failed to Get added item")
	}
	time.Sleep(ttl)
	if _, found := c.Get("foo"); found {
		t.Error("TTL cache did not expire entry after TTL")
	}
}
