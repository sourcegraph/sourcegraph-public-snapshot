package gitlab

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// paginatedResult provides an internal helper to manage paginated results from
// GitLab. (Surprise!)
//
// This is internal because exported APIs are generally going to need to return
// typed results. Client.GetMergeRequestNotes() is a good example of this in
// action.
//
// The GitLab documentation for pagination can be found at
// https://docs.gitlab.com/ee/api/README.html#pagination. At present, only
// offset based pagination is supported.
type paginatedResult struct {
	client         *Client
	method         string
	resultProvider func() interface{}
	url            string
}

// newPaginatedResult creates a new paginatedResult instance that can be used to
// manage a paginated GitLab result. method and url should be self explanatory;
// resultProvider needs to return an empty instance of the desired result type
// so that the payload can be unmarshalled.
func (c *Client) newPaginatedResult(method, url string, resultProvider func() interface{}) *paginatedResult {
	return &paginatedResult{
		client:         c,
		method:         method,
		resultProvider: resultProvider,
		url:            url,
	}
}

// next returns the next page in the result set. If there are no further pages,
// an empty result and a nil error are returned.
func (pr *paginatedResult) next(ctx context.Context) (interface{}, error) {
	// Special case: if the next page URL is a blank string, then there are no
	// more pages, and we can early return.
	if pr.url == "" {
		return pr.resultProvider(), nil
	}

	req, err := http.NewRequest(pr.method, pr.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	data := pr.resultProvider()
	header, _, err := pr.client.do(ctx, req, &data)
	if err != nil {
		return nil, errors.Wrap(err, "sending request")
	}
	pr.url = header.Get("X-Next-Page")

	return data, nil
}
