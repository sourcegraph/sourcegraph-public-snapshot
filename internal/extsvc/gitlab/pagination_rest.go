package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newRestPaginatedResult[T any](ctx context.Context, url string, c *Client) (*PaginatedResult[T], error) {
	pageNumber := 0

	return newPaginatedResult(func() ([]T, error) {
		var page []T

		pageUrl := fmt.Sprintf(url, pageNumber)
		pageNumber += 1
		req, err := http.NewRequest("GET", pageUrl, nil)
		if err != nil {
			return nil, errors.Wrap(err, "building request")
		}

		_, _, err = c.do(ctx, req, &page)
		if err != nil {
			return nil, errors.Wrap(err, "sending request")
		}

		return page, nil
	})
}
