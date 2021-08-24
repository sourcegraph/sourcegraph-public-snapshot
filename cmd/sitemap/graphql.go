package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"

	"github.com/cockroachdb/errors"
)

// This file contains all the methods required to execute Sourcegraph GraphQL API requests.

// graphQLQuery describes a general GraphQL query and its variables.
type graphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type graphQLClient struct {
	URL   string
	Token string
}

// requestGraphQL performs a GraphQL request with the given query and variables.
// search executes the given search query. The queryName is used as the source of the request.
// The result will be decoded into the given pointer.
func (c *graphQLClient) requestGraphQL(ctx context.Context, queryName string, query string, variables interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Encode")
	}

	req, err := http.NewRequest("POST", c.URL+"?"+queryName, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	var errs struct {
		Errors []interface{}
	}
	if err := json.Unmarshal(data, &errs); err != nil {
		return nil, errors.Wrap(err, "Unmarshal errors")
	}
	if len(errs.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %v", errs.Errors)
	}
	return data, nil
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
