package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

type streamClient struct {
	token    string
	endpoint string
	client   *http.Client
}

func newStreamClient() (*streamClient, error) {
	tkn := os.Getenv(envToken)
	if tkn == "" {
		return nil, fmt.Errorf("%s not set", envToken)
	}
	endpoint := os.Getenv(envEndpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("%s not set", envEndpoint)
	}

	return &streamClient{
		token:    tkn,
		endpoint: endpoint,
		client:   http.DefaultClient,
	}, nil
}

func (s *streamClient) search(ctx context.Context, query, queryName string) (*metrics, error) {
	req, err := streamhttp.NewRequest(s.endpoint, query)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("X-Sourcegraph-Should-Trace", "true")
	req.Header.Set("User-Agent", fmt.Sprintf("SearchBlitz (%s)", queryName))

	start := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := []streamhttp.EventMatch{}
	dec := streamhttp.Decoder{
		OnMatches: func(matches []streamhttp.EventMatch) {
			res = append(res, matches...)
		},
	}

	if err := dec.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	m := &metrics{
		took:  time.Since(start).Milliseconds(),
		trace: resp.Header.Get("x-trace"),
	}
	return m, nil
}

func (s *streamClient) clientType() string {
	return "stream"
}
