package graphqlbackend

import (
	"context"
	"io"
	"net/http"
	"sync"
)

type LinkMetadataResolver struct {
	once sync.Once
	URL  string

	// Cache results because they are used by multiple fields
	html string
}

type linkMetadataArgs struct {
	URL string
}

func (r *schemaResolver) LinkMetadata(ctx context.Context, args *linkMetadataArgs) *LinkMetadataResolver {
	return &LinkMetadataResolver{sync.Once{}, args.URL, ""}
}

func (r *LinkMetadataResolver) retrieveMetaTags(url string) string {
	r.once.Do(func() {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		r.html = string(body)

		// TODO: parse HTML, return map of meta tag name to value.
	})

	return r.html
}

func (r *LinkMetadataResolver) ImageURL() *string {
	html := r.retrieveMetaTags(r.URL)

	return strptr(html)
}

func (r *LinkMetadataResolver) Title() *string {
	html := r.retrieveMetaTags(r.URL)

	return strptr(html)
}

func (r *LinkMetadataResolver) Description() *string {
	html := r.retrieveMetaTags(r.URL)

	return strptr(html)
}
