// Package loggedcache provides a cache wrapper that logs and records metrics
// about operations performed on an underlying cache.
package loggedcache

import (
	"log"
	"time"
)

// Async is a cache wrapper that asynchronously logs and records metrics about
// synchronous operations performed on an underlying cache.
type Async struct {
	// Underlying is the underlying cache to forward cache operations to.
	Underlying interface {
		Get(key string) (responseBytes []byte, ok bool)
		Set(key string, responseBytes []byte)
		Delete(key string)
	}

	// Count is called with "get"/"set"/"delete" for each operation performed on
	// the cache, even if it is unsuccessful.
	Count func(operation string)

	// Hit is called on each Get operation that returns data (i.e., ok is
	// true).
	Hit func()

	// Time is called with "get"/"set"/"delete" and the total duration for each
	// operation performed on the cache.
	Time func(operation string, t time.Duration)

	// Log receives log messages for operations performed on the underlying cache.
	Log *log.Logger
}

func (c *Async) Get(key string) (resp []byte, ok bool) {
	if c.Count != nil {
		go c.Count("get")
	}
	var t0 time.Time
	if c.Time != nil {
		t0 = time.Now()
	}

	if c.Log != nil {
		c.Log.Printf("get %q", key)
	}

	resp, ok = c.Underlying.Get(key)

	if ok {
		if c.Hit != nil {
			go c.Hit()
		}
		if c.Log != nil {
			go c.Log.Printf("get hit %q", key)
		}
	} else {
		if c.Log != nil {
			go c.Log.Printf("get miss %q", key)
		}
	}

	if c.Time != nil {
		dt := time.Since(t0)
		go c.Time("get", dt)
		if c.Log != nil {
			go c.Log.Printf("get time %s [data: %d bytes]", dt, len(resp))
		}
	}

	return
}

func (c *Async) Set(key string, data []byte) {
	if c.Count != nil {
		go c.Count("set")
	}
	var t0 time.Time
	if c.Time != nil {
		t0 = time.Now()
	}

	if c.Log != nil {
		c.Log.Printf("set %q [data: %d bytes]", key, len(data))
	}

	c.Underlying.Set(key, data)

	if c.Time != nil {
		dt := time.Since(t0)
		go c.Time("set", dt)
		if c.Log != nil {
			go c.Log.Printf("set time %s [data: %d bytes]", dt, len(data))
		}
	}
}

func (c *Async) Delete(key string) {
	if c.Count != nil {
		go c.Count("delete")
	}
	var t0 time.Time
	if c.Time != nil {
		t0 = time.Now()
	}

	if c.Log != nil {
		c.Log.Printf("delete %q", key)
	}

	c.Underlying.Delete(key)

	if c.Time != nil {
		dt := time.Since(t0)
		go c.Time("delete", dt)
		if c.Log != nil {
			go c.Log.Printf("delete time %s", dt)
		}
	}
}
