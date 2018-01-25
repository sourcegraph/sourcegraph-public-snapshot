package repoupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

var repoupdaterURL = env.Get("REPO_UPDATER_URL", "http://repo-updater:3182", "repo-updater server URL")

var (
	// ErrNotFound is when a repository is not found.
	ErrNotFound = errors.New("repository not found")

	// ErrUnauthorized is when an authorization error occurred.
	ErrUnauthorized = errors.New("not authorized")
)

// DefaultClient is the default Client. Unless overwritten, it is connected to the server specified by the
// REPO_UPDATER_URL environment variable.
var DefaultClient = &Client{
	URL: repoupdaterURL,
	HTTPClient: &http.Client{
		// nethttp.Transport will propogate opentracing spans
		Transport: &nethttp.Transport{
			RoundTripper: &http.Transport{
				// Default is 2, but we can send many concurrent requests
				MaxIdleConnsPerHost: 500,
			},
		},
	},
}

// Client is a repoupdater client.
type Client struct {
	// URL to repoupdater server.
	URL string

	// HTTP client to use
	HTTPClient *http.Client
}

// RepoLookup retrieves information about the repository on repoupdater.
func (c *Client) RepoLookup(ctx context.Context, repo api.RepoURI) (result *protocol.RepoLookupResult, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Client.RepoLookup")
	defer func() {
		if result != nil {
			span.SetTag("found", result.Repo != nil)
		}
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Repo", string(repo))

	resp, err := c.httpPost(ctx, "repo-lookup", protocol.RepoLookupArgs{Repo: repo})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var err error
		if resp.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		} else if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			err = ErrUnauthorized
		} else {
			err = fmt.Errorf("http status %d", resp.StatusCode)
		}
		return nil, errors.Wrap(err, fmt.Sprintf("RepoLookup: %s", repo))
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (c *Client) httpPost(ctx context.Context, method string, payload interface{}) (resp *http.Response, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Client.httpPost")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.URL+"/"+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("RepoUpdater Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	if c.HTTPClient != nil {
		return c.HTTPClient.Do(req)
	}
	return http.DefaultClient.Do(req)
}
