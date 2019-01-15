package se

import (
	"context"
	"fmt"
	"net/url"
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
				t.Failed()
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
				t.Failed()
			}
		})
		t.Run(fmt.Sprintf("custom client %s", tc.urlStr), func(t *testing.T) {
			vals, allowed := Client{allowList: DefaultAllowListPatterns}.IsAllowedURL(tc.urlStr)
			if allowed != tc.allowed {
				t.Fatalf("Expected %t to equal %t when checking if %q is allowed.", allowed, tc.allowed, tc.urlStr)
				t.Failed()
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
				t.Failed()
			}
		})
	}
}

func Test_FetchUpdate(t *testing.T) {

	var c, _ = NewClient(SpecifyLockWaitTimeout(50 * time.Millisecond))

	t.Run("locking", func(t *testing.T) {
		t.Run("errs with a lock contnetion error on time", func(t *testing.T) {

			t.Skip("needs configurable workdir before I can write lockfiles to disk")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			go c.FetchUpdate(ctx, "http://stackoverflow.com/...")
			c.FetchUpdate(ctx, "http://stackoverflow.com/...")
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
