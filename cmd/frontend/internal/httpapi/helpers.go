package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// writeJSON writes a JSON Content-Type header and a JSON-encoded object to the
// http.ResponseWriter.
func writeJSON(w http.ResponseWriter, v interface{}) error {
	// Return "[]" instead of "null" if v is a nil slice.
	if reflect.TypeOf(v).Kind() == reflect.Slice && reflect.ValueOf(v).IsNil() {
		v = []interface{}{}
	}

	// MarshalIndent takes about 30-50% longer, which
	// significantly increases the time it takes to handle and return
	// large HTTP API responses.
	w.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(v)
}

// writePaginationHeader writes an HTTP Link header with links to the first,
// previous, next, and last pages of a paginated result set. This lets API
// clients page through results without needing to construct URLs on their own.
//
// The Link header for pagination follows the format of GitHub's. See
// https://developer.github.com/v3/#pagination for more information.
func writePaginationHeader(w http.ResponseWriter, current *url.URL, listOpts sourcegraph.ListOptions, total int) {
	page, perPage := int32(listOpts.PageOrDefault()), int32(listOpts.PerPageOrDefault())
	numPages := int32((int32(total) / perPage) + 1)

	type link struct {
		rel string
		url string
	}
	var links []link

	if page != 1 {
		links = append(links, link{
			rel: "first",
			url: urlWithListOptions(current, sourcegraph.ListOptions{Page: 1, PerPage: perPage}),
		})
		links = append(links, link{
			rel: "prev",
			url: urlWithListOptions(current, sourcegraph.ListOptions{Page: page - 1, PerPage: perPage}),
		})
	}
	if page != numPages {
		links = append(links, link{
			rel: "next",
			url: urlWithListOptions(current, sourcegraph.ListOptions{Page: page + 1, PerPage: perPage}),
		})
		links = append(links, link{
			rel: "last",
			url: urlWithListOptions(current, sourcegraph.ListOptions{Page: numPages, PerPage: perPage}),
		})
	}

	linkStrs := make([]string, len(links))
	for i, link := range links {
		linkStrs[i] = fmt.Sprintf(`<%s>; rel="%s"`, link.url, link.rel)
	}

	w.Header().Add("Link", strings.Join(linkStrs, ", "))
}

func urlWithListOptions(u *url.URL, listOpts sourcegraph.ListOptions) string {
	q := u.Query()
	qs, err := query.Values(listOpts)
	if err != nil {
		panic("query.Values: " + err.Error())
	}
	if listOpts.Page > 1 {
		q["Page"] = qs["Page"]
	} else {
		delete(q, "Page")
	}
	u.RawQuery = q.Encode()
	return u.String()
}
