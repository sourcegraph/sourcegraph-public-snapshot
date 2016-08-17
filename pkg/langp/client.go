package langp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
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
func (c *Client) Prepare(r *RepoRev) error {
	return c.do("prepare", r, nil)
}

// PositionToDefKey returns the DefKey for the given position.
func (c *Client) PositionToDefKey(p *Position) (*DefKey, error) {
	var result DefKey
	err := c.do("position-to-defkey", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DefKeyToPosition returns the position of the given DefKey.
func (c *Client) DefKeyToPosition(k *DefKey) (*Position, error) {
	var result Position
	err := c.do("defkey-to-position", k, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Definition resolves the specified position, effectively returning where the
// given definition is defined. For example, this is used for go to definition.
func (c *Client) Definition(p *Position) (*Range, error) {
	var result Range
	err := c.do("definition", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Hover returns hover-over information about the def/ref/etc at the given
// position.
func (c *Client) Hover(p *Position) (*Hover, error) {
	var result Hover
	err := c.do("hover", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// LocalRefs resolves references to repository-local definitions.
func (c *Client) LocalRefs(p *Position) (*RefLocations, error) {
	var result RefLocations
	err := c.do("local-refs", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExternalRefs resolves references to repository-external definitions.
func (c *Client) ExternalRefs(r *RepoRev) (*ExternalRefs, error) {
	var result ExternalRefs
	err := c.do("external-refs", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportedSymbols lists repository-local definitions which are exported.
func (c *Client) ExportedSymbols(r *RepoRev) (*ExportedSymbols, error) {
	var result ExportedSymbols
	err := c.do("exported-symbols", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) do(endpoint string, body, results interface{}) error {
	// TODO: maybe consider retrying upon first request failure to prevent
	// such errors from ending up on the frontend for reliability purposes.
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.endpoint(endpoint), bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp Error
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("error parsing language processor error (status code %v): %v", resp.StatusCode, err)
		}
		return &errResp
	}
	if results == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(results)
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
