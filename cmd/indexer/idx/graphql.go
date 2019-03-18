package idx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"golang.org/x/net/context/ctxhttp"
)

type graphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

const gqlLangServersEnableMutation = `mutation EnableLanguage($language: String!) {
  langServers {
    enable(language: $language) {
      alwaysNil
    }
  }
}
`

type gqlLangServersEnableVars struct {
	Language string `json:"language"`
}

type gqlResponse struct {
	Data   struct{}
	Errors []interface{}
}

func langServersEnableLanguage(ctx context.Context, language string) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     gqlLangServersEnableMutation,
		Variables: gqlLangServersEnableVars{Language: language},
	})
	if err != nil {
		return errors.Wrap(err, "Encode")
	}

	url, err := gqlURL("EnableLanguage")
	if err != nil {
		return errors.Wrap(err, "constructing frontend URL")
	}

	resp, err := ctxhttp.Post(ctx, nil, url, "application/json", &buf)
	if err != nil {
		return errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	var res *gqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return errors.Wrap(err, "Decode")
	}
	if len(res.Errors) > 0 {
		return fmt.Errorf("graphql: errors: %v", res.Errors)
	}
	return nil
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
