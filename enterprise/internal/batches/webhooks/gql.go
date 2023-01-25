package webhooks

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
)

func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}
