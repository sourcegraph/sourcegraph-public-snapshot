package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	envToken = "SOURCEGRAPH_TOKEN"
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

func (s *client) search(ctx context.Context, queryString string) (*result, *metrics, error) {
	var body bytes.Buffer
	m :=&metrics{}
	if err := json.NewEncoder(&body).Encode(map[string]interface{}{
		"query":     graphQLQuery,
		"variables": map[string]string{"query": queryString},
	}); err != nil {
		return nil, nil,  err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.url(), ioutil.NopCloser(&body))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", "token "+s.token)

	start := time.Now()
	resp, err := s.client.Do(req)
	m.took = time.Since(start).Milliseconds()

	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		break
	default:
		return nil, nil,fmt.Errorf("Unexpected status code", resp.StatusCode)
	}

	// Decode the response.
	respDec := rawResult{Data: result{}}
	if err := json.NewDecoder(resp.Body).Decode(&respDec); err != nil {
		return nil, nil, err
	}
	return &respDec.Data, m, nil
}

func (s *client) url() string {
	return s.endpoint + "/.api/graphql"
}
