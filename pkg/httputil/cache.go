package httputil

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/rcache"

	"github.com/gregjones/httpcache"
)

var (
	// Cache is a HTTP cache backed by Redis. The TTL of a week is a
	// balance between caching values for a useful amount of time versus
	// growing the cache too large.
	Cache = rcache.NewWithTTL("http", 604800)

	// CachingClient is an HTTP client that caches responses backed by
	// Redis (using Cache).
	CachingClient = &http.Client{Transport: httpcache.NewTransport(Cache)}
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_838(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
