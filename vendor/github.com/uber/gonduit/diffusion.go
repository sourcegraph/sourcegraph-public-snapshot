package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// DiffusionQueryCommitsMethod is the method name on API.
const DiffusionQueryCommitsMethod = "diffusion.querycommits"

// DiffusionQueryCommits performs a call to diffusion.querycommits.
func (c *Conn) DiffusionQueryCommits(
	req requests.DiffusionQueryCommitsRequest,
) (*responses.DiffusionQueryCommitsResponse, error) {
	var res responses.DiffusionQueryCommitsResponse

	if err := c.Call(DiffusionQueryCommitsMethod, &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// DiffusionRepositorySearchMethod is the method name on API.
const DiffusionRepositorySearchMethod = "diffusion.repository.search"

// DiffusionRepositorySearch calls "diffusion.repository.search" Conduit API
// method.
func (c *Conn) DiffusionRepositorySearch(
	req requests.DiffusionRepositorySearchRequest,
) (*responses.DiffusionRepositorySearchResponse, error) {
	var resp responses.DiffusionRepositorySearchResponse
	if err := c.Call(DiffusionRepositorySearchMethod, &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
