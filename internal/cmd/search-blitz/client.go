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
		return nil, fmt.Errorf("%s not set", envToken)
	}
	endpoint := os.Getenv(envEndpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("%s not set", envEndpoint)
	}

	return &client{
		token:    tkn,
		endpoint: endpoint,
		client:   http.DefaultClient,
	}, nil
}

func (s *client) search(ctx context.Context, query, queryName string) (*metrics, error) {
	var body bytes.Buffer
	m := &metrics{}
	if err := json.NewEncoder(&body).Encode(map[string]interface{}{
		"query":     graphQLQuery,
		"variables": map[string]string{"query": query},
	}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.url(), io.NopCloser(&body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("X-Sourcegraph-Should-Trace", "true")
	req.Header.Set("User-Agent", fmt.Sprintf("SearchBlitz (%s)", queryName))

	start := time.Now()
	resp, err := s.client.Do(req)
	m.took = time.Since(start).Milliseconds()

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		break
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	m.trace = resp.Header.Get("x-trace")

	// Decode the response.
	respDec := rawResult{Data: result{}}
	if err := json.NewDecoder(resp.Body).Decode(&respDec); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *client) url() string {
	return s.endpoint + "/.api/graphql?SearchBlitz"
}

func (s *client) clientType() string {
	return "batch"
}
