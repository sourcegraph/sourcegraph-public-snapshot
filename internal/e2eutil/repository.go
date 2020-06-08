package e2eutil

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// WaitForReposToBeCloned waits (up to 30 seconds) for all repositories
// in the list to be cloned.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeCloned(repos ...string) error {
	return retry(30*time.Second, func() error {
		const query = `
query Repositories($first: Int) {
	repositories(first: $first, cloned: true) {
		nodes {
			name
		}
	}
}
`
		variables := map[string]interface{}{
			"first": len(repos),
		}
		var resp struct {
			Data struct {
				Repositories struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"repositories"`
			} `json:"data"`
		}
		err := c.GraphQL("", query, variables, &resp)
		if err != nil {
			return errors.Wrap(err, "request GraphQL")
		}

		if len(resp.Data.Repositories.Nodes) != len(repos) {
			return errContinueRetry
		}

		repoSet := make(map[string]struct{}, len(repos))
		for _, repo := range repos {
			repoSet[repo] = struct{}{}
		}
		for _, node := range resp.Data.Repositories.Nodes {
			delete(repoSet, node.Name)
		}
		if len(repoSet) > 0 {
			return errContinueRetry
		}

		return nil
	})
}

var errContinueRetry = errors.New("continue retry")

// retry retries the given function until reached timeout. The function should
// return errContinueRetry to indicate another retry.
func retry(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return errors.Errorf("retry timed out in %s", timeout)
			}
			return ctx.Err()
		default:
			err := fn()
			if err != errContinueRetry {
				return err
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
