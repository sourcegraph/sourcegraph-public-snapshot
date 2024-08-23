package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// RepositoryQuery performs a call to repository.query.
func (c *Conn) RepositoryQuery(
	req requests.RepositoryQueryRequest,
) (*responses.RepositoryQueryResponse, error) {
	var res responses.RepositoryQueryResponse

	if err := c.Call("repository.query", &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
