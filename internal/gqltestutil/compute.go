package gqltestutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ComputeResult struct {
	Repository
	MatchCount map[string]int
}

type ComputeMatchContext struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Matches []struct {
		Value       string `json:"value"`
		Environment struct {
			Variable string `json:"variable"`
			Value    string `json:"value"`
		} `json:"environment"`
	} `json:"matches"`
}

type ComputeStreamClient struct {
	*Client
}

// Compute runs a Compute search and transforms the GraphQL results into our internal representation
// of compute matches.
func (c *Client) Compute(query string) ([]*ComputeResult, error) {
	const gqlQuery = `
query Run($query: String!) {
	compute(query: $query) {
		__typename
		... on ComputeMatchContext {
		repository {
			name
			// id
		}
		// commit
		// path
		matches {
			value
			environment {
			variable
			value
			}
		}
		}
	}
	}
`
	variables := map[string]any{
		"query": query,
	}
	var resp struct {
		Data struct {
			Compute []ComputeMatchContext `json:"compute"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "compute graphQL")
	}
	fmt.Println(resp)
	return nil, nil
}

// Compute runs a Compute streaming search and transforms the results into our internal
// representation of compute matches.
func (c *ComputeStreamClient) Compute(query string) ([]*ComputeResult, error) {
	var results []*ComputeResult
	var errs []string
	var alerts []string
	err := c.compute(query, ComputeMatchContextStreamDecoder{
		OnResult: func(results []MatchContext) {
			println(results)
		},
		OnError: func(eventError *streamhttp.EventError) {
			errs = append(errs, eventError.Message)
		},
		OnAlert: func(eventAlert *streamhttp.EventAlert) {
			alerts = append(alerts, eventAlert.Description)
		},
		OnUnknown: func(event, data []byte) {
			// TODO: remove.
			println(string(data))
		},
	})
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, errors.Newf("compute streaming error: %v", errs)
	}
	return results, nil
}

// We copy the internal client here to use with public clients. Could maybe be renamed.
type ComputeMatchContextStreamDecoder struct {
	OnResult  func(results []MatchContext)
	OnAlert   func(*streamhttp.EventAlert)
	OnError   func(*streamhttp.EventError)
	OnUnknown func(event, data []byte)
}

type Match struct {
	Value string `json:"value"`
}

type MatchContext struct {
	Matches      []Match `json:"matches"`
	Path         string  `json:"path"`
	RepositoryID int32   `json:"repositoryID"`
	Repository   string  `json:"repository"`
}

// NewMatchContextRequest returns an http.Request against the streaming API for query.
func NewMatchContextRequest(baseURL string, query string) (*http.Request, error) {
	u := baseURL + "/compute/stream?q=" + url.QueryEscape(query)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

func (rr ComputeMatchContextStreamDecoder) ReadAll(r io.Reader) error {
	dec := streamhttp.NewDecoder(r)

	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()

		if bytes.Equal(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			var d []MatchContext
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode compute match context payload: %w", err)
			}
			rr.OnResult(d)
		} else if bytes.Equal(event, []byte("alert")) {
			if rr.OnAlert == nil {
				continue
			}
			var d streamhttp.EventAlert
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode alert payload: %w", err)
			}
			rr.OnAlert(&d)
		} else if bytes.Equal(event, []byte("error")) {
			if rr.OnError == nil {
				continue
			}
			var d streamhttp.EventError
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode error payload: %w", err)
			}
			rr.OnError(&d)
		} else if bytes.Equal(event, []byte("done")) {
			// Always the last event
			break
		} else {
			if rr.OnUnknown == nil {
				continue
			}
			rr.OnUnknown(event, data)
		}
	}
	return dec.Err()
}

func (c *ComputeStreamClient) compute(query string, dec ComputeMatchContextStreamDecoder) error {
	req, err := streamhttp.NewRequest(strings.TrimRight(c.Client.baseURL, "/")+"/.api", query)
	if err != nil {
		return err
	}
	// Note: Sending this header enables us to use session cookie auth without sending a trusted Origin header.
	// https://docs.sourcegraph.com/dev/security/csrf_security_model#authentication-in-api-endpoints
	req.Header.Set("X-Requested-With", "Sourcegraph")
	c.Client.addCookies(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return dec.ReadAll(resp.Body)
}
