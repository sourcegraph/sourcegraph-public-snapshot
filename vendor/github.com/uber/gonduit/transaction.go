package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

const TransactionSearchMethod = "transaction.search"

// TransactionSearch performs a call to transaction.search.
func (c *Conn) TransactionSearch(
	req requests.TransactionSearchRequest,
) (*responses.TransactionSearchResponse, error) {
	var res responses.TransactionSearchResponse

	if err := c.Call(TransactionSearchMethod, &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
