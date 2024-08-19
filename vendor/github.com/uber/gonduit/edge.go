package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

const EdgeSearchMethod = "edge.search"

// EdgeSearch performs a call to edge.search.
func (c *Conn) EdgeSearch(
	req requests.EdgeSearchRequest,
) (*responses.EdgeSearchResponse, error) {
	var res responses.EdgeSearchResponse

	if err := c.Call(EdgeSearchMethod, &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
