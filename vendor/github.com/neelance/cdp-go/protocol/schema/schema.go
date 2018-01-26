// Provides information about the protocol schema.
package schema

import (
	"github.com/neelance/cdp-go/rpc"
)

// Provides information about the protocol schema.
type Client struct {
	*rpc.Client
}

// Description of the protocol domain.

type Domain struct {
	// Domain name.
	Name string `json:"name"`

	// Domain version.
	Version string `json:"version"`
}

type GetDomainsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns supported domains.
func (d *Client) GetDomains() *GetDomainsRequest {
	return &GetDomainsRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetDomainsResult struct {
	// List of supported domains.
	Domains []*Domain `json:"domains"`
}

func (r *GetDomainsRequest) Do() (*GetDomainsResult, error) {
	var result GetDomainsResult
	err := r.client.Call("Schema.getDomains", r.opts, &result)
	return &result, err
}

func init() {
}
