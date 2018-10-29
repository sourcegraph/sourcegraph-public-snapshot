package idx

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type graphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(api.InternalClient.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}
