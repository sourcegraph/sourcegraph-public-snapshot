package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This file contains all the methods required to execute Sourcegraph GraphQL API requests.

var (
	graphQLTimeout, _          = time.ParseDuration(env.Get("GRAPHQL_TIMEOUT", "30s", "Timeout for GraphQL HTTP requests"))
	graphQLRetryDelayBase, _   = time.ParseDuration(env.Get("GRAPHQL_RETRY_DELAY_BASE", "200ms", "Base retry delay duration for GraphQL HTTP requests"))
	graphQLRetryDelayMax, _    = time.ParseDuration(env.Get("GRAPHQL_RETRY_DELAY_MAX", "3s", "Max retry delay duration for GraphQL HTTP requests"))
	graphQLRetryMaxAttempts, _ = strconv.Atoi(env.Get("GRAPHQL_RETRY_MAX_ATTEMPTS", "20", "Max retry attempts for GraphQL HTTP requests"))
)

// graphQLQuery describes a general GraphQL query and its variables.
type graphQLQuery struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

type graphQLClient struct {
	URL   string
	Token string

	factory *httpcli.Factory
}

// requestGraphQL performs a GraphQL request with the given query and variables.
// search executes the given search query. The queryName is used as the source of the request.
// The result will be decoded into the given pointer.
func (c *graphQLClient) requestGraphQL(ctx context.Context, queryName string, query string, variables any) ([]byte, error) {
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

	if c.factory == nil {
		c.factory = httpcli.NewFactory(
			httpcli.NewMiddleware(
				httpcli.ContextErrorMiddleware,
			),
			httpcli.NewMaxIdleConnsPerHostOpt(500),
			httpcli.NewTimeoutOpt(graphQLTimeout),
			// ExternalTransportOpt needs to be before TracedTransportOpt and
			// NewCachedTransportOpt since it wants to extract a http.Transport,
			// not a generic http.RoundTripper.
			httpcli.ExternalTransportOpt,
			httpcli.NewErrorResilientTransportOpt(
				httpcli.NewRetryPolicy(httpcli.MaxRetries(graphQLRetryMaxAttempts)),
				httpcli.ExpJitterDelay(graphQLRetryDelayBase, graphQLRetryDelayMax),
			),
			httpcli.TracedTransportOpt,
		)
	}

	doer, err := c.factory.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "Doer")
	}
	resp, err := doer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	var errs struct {
		Errors []any
	}
	if err := json.Unmarshal(data, &errs); err != nil {
		return nil, errors.Wrap(err, "Unmarshal errors")
	}
	if len(errs.Errors) > 0 {
		return nil, errors.Newf("graphql error: %v", errs.Errors)
	}
	return data, nil
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
