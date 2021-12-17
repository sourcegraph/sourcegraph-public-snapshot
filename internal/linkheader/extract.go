package linkheader

import (
	"net/http"

	"github.com/tomnomnom/linkheader"
)

// ExtractNextURL retrieves the URL with rel="next" from the Link header.
func ExtractNextURL(resp *http.Response) (string, bool) {
	return ExtractURL(resp, "next")
}

// ExtractURL retrieves the URL with given rel froim the Link header.
func ExtractURL(resp *http.Response, rel string) (string, bool) {
	for _, link := range linkheader.Parse(resp.Header.Get("Link")) {
		if link.Rel == rel {
			return link.URL, true
		}
	}

	return "", false
}
