package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	envToken    = "SOURCEGRAPH_TOKEN"
	envEndpoint = "SOURCEGRAPH_ENDPOINT"
)

type client struct {
	token    string
	endpoint string
	client   *http.Client
}

func newClient() (*client, error) {
	tkn := os.Getenv(envToken)
	if tkn == "" {
		return nil, errors.Errorf("%s not set", envToken)
	}
	endpoint := os.Getenv(envEndpoint)
	if endpoint == "" {
		return nil, errors.Errorf("%s not set", envEndpoint)
	}

	return &client{
		token:    tkn,
		endpoint: endpoint,
		client:   http.DefaultClient,
	}, nil
}

func (s *client) search(ctx context.Context, query, queryName string) (*metrics, error) {
	return s.doGraphQL(ctx, graphQLRequest{
		QueryName:        queryName,
		GraphQLQuery:     graphQLSearchQuery,
		GraphQLVariables: map[string]string{"query": query},
		MetricsFromBody: func(body io.Reader) (*metrics, error) {
			var respDec struct {
				Data struct {
					Search struct{ Results struct{ MatchCount int } }
				}
			}
			if err := json.NewDecoder(body).Decode(&respDec); err != nil {
				return nil, err
			}
			return &metrics{
				matchCount: respDec.Data.Search.Results.MatchCount,
			}, nil
		},
	})
}

func (s *client) attribution(ctx context.Context, snippet, queryName string) (*metrics, error) {
	return s.doGraphQL(ctx, graphQLRequest{
		QueryName:        queryName,
		GraphQLQuery:     graphQLAttributionQuery,
		GraphQLVariables: map[string]string{"snippet": snippet},
		MetricsFromBody: func(body io.Reader) (*metrics, error) {
			var respDec struct {
				Data struct{ SnippetAttribution struct{ TotalCount int } }
			}
			if err := json.NewDecoder(body).Decode(&respDec); err != nil {
				return nil, err
			}
			return &metrics{
				matchCount: respDec.Data.SnippetAttribution.TotalCount,
			}, nil
		},
	})
}

type graphQLRequest struct {
	QueryName string

	GraphQLQuery     string
	GraphQLVariables map[string]string

	MetricsFromBody func(io.Reader) (*metrics, error)
}

func (s *client) doGraphQL(ctx context.Context, greq graphQLRequest) (*metrics, error) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]any{
		"query":     greq.GraphQLQuery,
		"variables": greq.GraphQLVariables,
	}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.url(), io.NopCloser(&body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("X-Sourcegraph-Should-Trace", "true")
	req.Header.Set("User-Agent", fmt.Sprintf("SearchBlitz (%s)", greq.QueryName))

	start := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		break
	default:
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response.
	metrics, err := greq.MetricsFromBody(resp.Body)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	metrics.took = duration
	metrics.firstResult = duration
	metrics.trace = resp.Header.Get("x-trace")

	return metrics, nil
}

func (s *client) url() string {
	return s.endpoint + "/.api/graphql?SearchBlitz"
}

func (s *client) clientType() string {
	return "batch"
}
