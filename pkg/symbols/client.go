package symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/endpoint"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
)

var symbolsURL = env.Get("SYMBOLS_URL", "http://symbols:3184", "symbols service URL")

// DefaultClient is the default Client. Unless overwritten, it is connected to the server specified by the
// SYMBOLS_URL environment variable.
var DefaultClient = &Client{
	URL: symbolsURL,
	HTTPClient: &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{
			RoundTripper: &http.Transport{
				// Default is 2, but we can send many concurrent requests
				MaxIdleConnsPerHost: 500,
			},
		},
	},
	HTTPLimiter: parallel.NewRun(500),
}

// Client is a symbols service client.
type Client struct {
	// URL to symbols service.
	URL string

	// HTTP client to use
	HTTPClient *http.Client

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run

	mu       sync.Mutex
	endpoint *endpoint.Map
}

type key struct {
	repo     api.RepoURI
	commitID api.CommitID
}

func (c *Client) url(key key) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.endpoint == nil {
		var err error
		c.endpoint, err = endpoint.New(c.URL)
		if err != nil {
			return "", err
		}
	}
	return c.endpoint.Get(string(key.repo) + ":" + string(key.commitID))
}

// Search performs a symbol search on the symbols service.
func (c *Client) Search(ctx context.Context, args protocol.SearchArgs) (result *protocol.SearchResult, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Client.Search")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Repo", string(args.Repo))
	span.SetTag("CommitID", string(args.CommitID))

	resp, err := c.httpPost(ctx, "search", key{repo: args.Repo, commitID: args.CommitID}, args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	stack := fmt.Sprintf("Search: %+v", args)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(fmt.Errorf("http status %d", resp.StatusCode), stack)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (c *Client) httpPost(ctx context.Context, method string, key key, payload interface{}) (resp *http.Response, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Client.httpPost")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	url, err := c.url(key)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	req, err := http.NewRequest("POST", url+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	if c.HTTPLimiter != nil {
		span.LogKV("event", "Waiting on HTTP limiter")
		c.HTTPLimiter.Acquire()
		defer c.HTTPLimiter.Release()
		span.LogKV("event", "Acquired HTTP limiter")
	}

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Symbols Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	if c.HTTPClient != nil {
		return c.HTTPClient.Do(req)
	}
	return http.DefaultClient.Do(req)
}
