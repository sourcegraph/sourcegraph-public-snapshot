package gonduit

import (
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// PHIDLookup calls the phid.lookup endpoint.
func (c *Conn) PHIDLookup(
	req requests.PHIDLookupRequest,
) (responses.PHIDLookupResponse, error) {
	var r responses.PHIDLookupResponse

	if err := c.Call("phid.lookup", &req, &r); err != nil {
		return nil, err
	}

	return r, nil
}

// PHIDLookupSingle calls the phid.lookup endpoint with a single name.
func (c *Conn) PHIDLookupSingle(name string) (*entities.PHIDResult, error) {
	req := requests.PHIDLookupRequest{
		Names: []string{name},
	}

	resp, err := c.PHIDLookup(req)

	if err != nil {
		return nil, err
	}

	return resp[name], nil
}

// PHIDQuery calls the phid.query endpoint.
func (c *Conn) PHIDQuery(
	req requests.PHIDQueryRequest,
) (responses.PHIDQueryResponse, error) {
	var r responses.PHIDQueryResponse

	if err := c.Call("phid.query", &req, &r); err != nil {
		return nil, err
	}

	return r, nil
}

// PHIDQuerySingle calls the phid.query endpoint with a single phid.
func (c *Conn) PHIDQuerySingle(phid string) (*entities.PHIDResult, error) {
	resp, err := c.PHIDQuery(requests.PHIDQueryRequest{
		PHIDs: []string{phid},
	})

	if err != nil {
		return nil, err
	}

	return resp[phid], nil
}
