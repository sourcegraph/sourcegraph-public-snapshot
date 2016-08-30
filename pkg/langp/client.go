package langp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

// Client is a Language Processor REST API client which is safe for use by
// multiple goroutines concurrently.
type Client struct {
	// Endpoint is the HTTP endpoint of the Language Processor.
	Endpoint *url.URL

	// Client, if specified, is used for making HTTP requests.
	Client *http.Client
}

// Prepare informs the Language Processor that it should prepare a workspace
// for the specified repo / commit. It is sent prior to an actual user request
// (e.g. as soon as we have access to their repos) in hopes of having
// preparation completed already when a user makes their first request.
func (c *Client) Prepare(ctx context.Context, r *RepoRev) error {
	return c.do(ctx, "prepare", r, nil)
}

// PositionToDefSpec returns the DefSpec for the given position.
func (c *Client) PositionToDefSpec(ctx context.Context, p *Position) (*DefSpec, error) {
	var result DefSpec
	err := c.do(ctx, "position-to-defspec", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DefSpecToPosition returns the position of the given DefSpec.
func (c *Client) DefSpecToPosition(ctx context.Context, k *DefSpec) (*Position, error) {
	var result Position
	err := c.do(ctx, "defspec-to-position", k, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Definition resolves the specified position, effectively returning where the
// given definition is defined. For example, this is used for go to definition.
func (c *Client) Definition(ctx context.Context, p *Position) (*Range, error) {
	var result Range
	err := c.do(ctx, "definition", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Hover returns hover-over information about the def/ref/etc at the given
// position.
func (c *Client) Hover(ctx context.Context, p *Position) (*Hover, error) {
	var result Hover
	err := c.do(ctx, "hover", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// LocalRefs resolves references to repository-local definitions.
func (c *Client) LocalRefs(ctx context.Context, p *Position) (*RefLocations, error) {
	var result RefLocations
	err := c.do(ctx, "local-refs", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DefSpecRefs resolves references to repository definitions.
func (c *Client) DefSpecRefs(ctx context.Context, k *DefSpec) (*RefLocations, error) {
	var result RefLocations
	err := c.do(ctx, "defspec-refs", k, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExternalRefs resolves references to repository-external definitions.
func (c *Client) ExternalRefs(ctx context.Context, r *RepoRev) (*ExternalRefs, error) {
	var result ExternalRefs
	err := c.do(ctx, "external-refs", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportedSymbols lists repository-local definitions which are exported.
func (c *Client) ExportedSymbols(ctx context.Context, r *RepoRev) (*ExportedSymbols, error) {
	var result ExportedSymbols
	err := c.do(ctx, "exported-symbols", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) do(ctx context.Context, endpoint string, body, results interface{}) error {
	// TODO: maybe consider retrying upon first request failure to prevent
	// such errors from ending up on the frontend for reliability purposes.
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint(endpoint), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s (body '%s')", err, string(data))
	}

	req.Header.Add("Content-Type", "application/json")

	operationName := fmt.Sprintf("LP Client: POST %s", c.endpoint(endpoint))
	span := opentracing.StartSpan(operationName, opentracing.ChildOf(opentracing.SpanFromContext(ctx).Context()))
	span.LogEventWithPayload("request body", body)
	defer span.Finish()

	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		return fmt.Errorf("%s (body '%s')", err, string(data))
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("%s (body '%s')", err, string(data))
	}
	defer resp.Body.Close()

	// 1 KB is a good, safe choice for medium-to-high throughput traces.
	saver := &prefixSuffixSaver{N: 1 * 1024}
	tee := io.TeeReader(resp.Body, saver)
	defer func() {
		span.LogEventWithPayload("response - "+resp.Status, string(saver.Bytes()))
	}()

	if resp.StatusCode != http.StatusOK {
		var errResp Error
		if err := json.NewDecoder(tee).Decode(&errResp); err != nil {
			return fmt.Errorf("error parsing language processor error (status code %v): %v", resp.StatusCode, err)
		}
		return &errResp
	}
	if results == nil {
		return nil
	}
	return json.NewDecoder(tee).Decode(results)
}

// endpoint returns a URL based on c.Endpoint with the given path suffixed.
func (c *Client) endpoint(p string) string {
	cpy := *c.Endpoint
	cpy.Path = path.Join(cpy.Path, p)
	return cpy.String()
}

// NewClient returns a new client with the default options connecting to the
// given Language Processor endpoint.
//
// An error is returned only if parsing the endpoint URL fails.
func NewClient(endpoint string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("must specify endpoint scheme")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("must specify endpoint host")
	}
	return &Client{
		Endpoint: u,
		Client: &http.Client{
			// TODO(slimsag): Once we have proper async operations we should
			// lower this timeout to respect those numbers. Until then, some
			// operations (listing all refs, cloning workspaces, etc) can take
			// quite a while and we don't want to abort the request.
			Timeout: 60 * time.Second,
		},
	}, nil
}
