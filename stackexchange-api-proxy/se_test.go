package se

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_IsAllowedURL(t *testing.T) {

	t.Parallel()

	var cases = []struct {
		urlStr  string
		vals    *url.Values
		allowed bool
	}{
		{
			urlStr:  "http://example.com",
			vals:    nil,
			allowed: false,
		},
		{
			urlStr:  "http://stackoverflow.com",
			vals:    &url.Values{"site": []string{"stackoverflow"}},
			allowed: true,
		},
		{
			urlStr:  "http://www.stackoverflow.com",
			vals:    &url.Values{"site": []string{"stackoverflow"}},
			allowed: true,
		},
		{
			urlStr:  "http://maliciousstackoverflow.com",
			vals:    nil,
			allowed: false,
		},
		{
			urlStr:  "http://www.stackexchange.com",
			vals:    nil,
			allowed: false,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("simple client %s", tc.urlStr), func(t *testing.T) {
			vals, allowed := IsAllowedURL(tc.urlStr)
			if allowed != tc.allowed {
				t.Fatalf("Expected %t to equal %t when checking if %q is allowed.", allowed, tc.allowed, tc.urlStr)
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
			}
		})
		t.Run(fmt.Sprintf("custom client %s", tc.urlStr), func(t *testing.T) {
			vals, allowed := Client{allowList: DefaultAllowListPatterns}.IsAllowedURL(tc.urlStr)
			if allowed != tc.allowed {
				t.Fatalf("Expected %t to equal %t when checking if %q is allowed.", allowed, tc.allowed, tc.urlStr)
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
			}
		})
	}
}

func Test_FetchUpdate(t *testing.T) {

	t.Run("allow-list", func(t *testing.T) {
		var c, _ = NewClient(SpecifyAllowList(AllowList{})) // empty allow-list
		err := c.FetchUpdate(context.Background(), "http://www.example.com/")
		if err != ErrURLNotAllowed {
			t.Fatalf("Expected all urls to be disallowed and return ErrURLNotAllowed, got %q", err)
		}
	})

	t.Run("unparsable URL", func(t *testing.T) {
		t.Skip("can't find anything that Go won't parse!")
		var c, _ = NewClient(SpecifyAllowList(AllowList{"helloworld": regexp.MustCompile(".*")}))
		err := c.FetchUpdate(context.Background(), "hello::::world") // non-integer ports are disallowed by http spec
		if err == nil {
			t.Fatalf("Expected string to return an error propagated from url.Parse, got nil")
		}
		if seErr, ok := err.(Error); !ok {
			t.Fatalf("Expected string to return an error propagated from url.Parse, got nil")
		} else {
			if seErr.Op != "parse-url" {
				t.Fatalf("Expected string to return an error propagated from url.Parse with .Op set to 'parse-url'")
			}
		}
	})

	t.Run("locking", func(t *testing.T) {
		t.Run("errs with a lock contention error on time", func(t *testing.T) {

			// Pre-lock a mutex, guarantees no race conditions
			// in trying to get and hold a lock whilst firing a second
			// request to the same URL resource.
			var alreadyLockedMutex = &sync.Mutex{}
			alreadyLockedMutex.Lock()
			defer alreadyLockedMutex.Unlock()

			var artificallyShortLockWaitTimeout = 50 * time.Millisecond
			var c, _ = NewClient(
				SpecifyLockWaitTimeout(artificallyShortLockWaitTimeout),
				SpecifyLockMechanism(func(u url.URL) sync.Locker {
					return alreadyLockedMutex
				}),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// start one we care about, and time it..
			start := time.Now()
			err := c.FetchUpdate(ctx, "http://stackoverflow.com/...")
			elapsed := time.Since(start)

			if err != ErrTimeoutLocking {
				t.Fatalf("Expected err to be %s got %s", ErrTimeoutLocking, err)
			}

			// Grace period of double the wait timeout. Observed was 5-7ms
			// overhead through, presumably the locking and scheduling of
			// goroutines and the use of select{}
			if elapsed > artificallyShortLockWaitTimeout*2 {
				t.Fatalf("Lock wait time was configured at %s but elapsed time was %s, must be a mistake in the locking code", artificallyShortLockWaitTimeout, elapsed)
			}

		})
	})

}

// TODO: This test could be much more comprehensive, this is just a
// big sanity check.
func Test_NewClient(t *testing.T) {
	var c1, _ = NewClient()
	var c2, _ = NewClient(SpecifyLockWaitTimeout(c1.lockWaitTimeout * 2))
	if c1.lockWaitTimeout == c2.lockWaitTimeout {
		t.Fatalf("Expected %s to equal %s when checking whether client optionsFns were applied correctly.", c1.lockWaitTimeout, c2.lockWaitTimeout)
	}
}
