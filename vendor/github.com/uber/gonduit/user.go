package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// UserQuery performs a call to differential.query.
func (c *Conn) UserQuery(
	req requests.UserQueryRequest,
) (*responses.UserQueryResponse, error) {
	var res responses.UserQueryResponse

	if err := c.Call("user.query", &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
