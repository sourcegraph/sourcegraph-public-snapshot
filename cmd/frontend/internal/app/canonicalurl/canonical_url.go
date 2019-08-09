// Package canonicalurl creates canonical URLs from request URLs by
// stripping extraneous query params, etc.
package canonicalurl

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/returnto"
)

// nonCanonicalQueryParams are query parameters that do not affect
// what's displayed on our site.
var nonCanonicalQueryParams = []string{
	"utm_source", "utm_medium", "utm_campaign", returnto.ParamName,
}

// FromURL returns the canonical URL for the given URL. The given URL
// is not modified.
func FromURL(currentURL *url.URL) *url.URL {
	canonicalQuery := currentURL.Query()
	for _, k := range nonCanonicalQueryParams {
		canonicalQuery.Del(k)
	}
	canonicalURL := *currentURL
	canonicalURL.RawQuery = canonicalQuery.Encode()
	return &canonicalURL
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_256(size int) error {
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
